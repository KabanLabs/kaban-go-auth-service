package pg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/VACdotCS/kaban-go-auth-service/internal/domain/models"
	"github.com/VACdotCS/kaban-go-auth-service/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, dbUrl string) (*Storage, error) {
	config, err := pgxpool.ParseConfig(dbUrl)

	if err != nil {
		return nil, err
	}

	pool, connErr := pgxpool.NewWithConfig(ctx, config)

	if connErr != nil {
		return nil, err
	}

	err = pool.Ping(ctx)

	if err != nil {
		return nil, err
	}

	return &Storage{Pool: pool}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (uid string, err error) {
	const op = "storage.pg.saveUser"

	err = s.Pool.QueryRow(ctx,
		"INSERT INTO users(email, pass_hash) VALUES ($1, $2) RETURNING id",
		email, passHash).Scan(&uid)

	if err != nil {
		if isUniqueViolation(err) {
			return "", fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return uid, nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "storage.pg.User"
	var user models.User

	err := s.Pool.QueryRow(ctx,
		"SELECT id, email, pass_hash, created_at, updated_at FROM users WHERE email = $1",
		email).Scan(&user.ID, &user.Email, &user.PassHash, &user.Created, &user.Updated)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) UserById(ctx context.Context, uid string) (models.User, error) {
	const op = "storage.pg.UserById"
	var user models.User

	err := s.Pool.QueryRow(ctx,
		"SELECT id, email, pass_hash, created_at, updated_at FROM users WHERE id = $1",
		uid).Scan(&user.ID, &user.Email, &user.PassHash, &user.Created, &user.Updated)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) App(ctx context.Context, appID int) (models.App, error) {
	const op = "storage.pg.App"
	var app models.App

	err := s.Pool.QueryRow(ctx,
		"SELECT id, title, scopes FROM apps WHERE id = $1",
		appID).Scan(&app.ID, &app.Name, &app.Scopes)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}

func (s *Storage) RefreshToken(ctx context.Context, refreshToken string) (models.RefreshToken, error) {
	const op = "storage.pg.RefreshToken"
	var token models.RefreshToken

	q := `
	SELECT 
		id, 
		token, 
		app_id, 
		user_id, 
		expires_at, 
		rotated, 
		created_at 
	FROM users_tokens 
	WHERE token =$1 AND rotated = false`

	err := s.Pool.QueryRow(ctx, q, refreshToken).
		Scan(
			&token.ID,
			&token.Token,
			&token.AppID,
			&token.UserID,
			&token.ExpireAt,
			&token.Rotated,
			&token.CreatedAt,
		)

	if err != nil {
		// TODO: Add another error if key was rotated
		if errors.Is(err, pgx.ErrNoRows) {
			return models.RefreshToken{}, fmt.Errorf("%s: %w", op, storage.ErrRefreshTokenNotFound)
		}
		return models.RefreshToken{}, fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (s *Storage) SaveRefreshToken(ctx context.Context, refreshToken string, uid string, expAt time.Time, appId int) (*models.RefreshToken, error) {
	const op = "storage.pg.SaveRefreshToken"
	var token models.RefreshToken

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: begin tx failed: %w", op, err)
	}
	defer tx.Rollback(ctx)

	rotateLastTokenQuery := `
		WITH last_token AS (
			SELECT id FROM users_tokens
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT 1
		)
		UPDATE users_tokens
		SET rotated = true
		WHERE id = (SELECT id FROM last_token)
	`
	cmdTag, err := tx.Exec(ctx, rotateLastTokenQuery, uid)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to rotate last token: %w", op, err)
	}

	if cmdTag.RowsAffected() == 0 {
		//
	}

	insertQuery := `
		INSERT INTO users_tokens (user_id, token, app_id, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	err = tx.QueryRow(ctx, insertQuery, uid, refreshToken, appId, expAt).Scan(&token.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to insert new token: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: commit failed: %w", op, err)
	}

	token.UserID = uid
	token.Token = refreshToken

	return &token, nil
}

func (s *Storage) RotateRefreshToken(ctx context.Context, refreshToken string) error {
	const op = "storage.pg.RotateRefreshToken"

	const rotateTokenQuery = "UPDATE users_tokens SET rotated = true WHERE token = $1"

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: begin tx failed: %w", op, err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, rotateTokenQuery, refreshToken)

	if err != nil {
		return fmt.Errorf("%s: exec query failed: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: commit failed: %w", op, err)
	}

	return nil
}

// IsUniqueViolation check is pgx error is about unique constraint violation
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

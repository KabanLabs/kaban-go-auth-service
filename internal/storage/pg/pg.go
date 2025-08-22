package pg

import (
	"context"
	"errors"
	"fmt"

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
		panic(connErr)
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
		email).Scan(&user.ID, &user.Email, &user.Created, &user.Updated)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) App(ctx context.Context, appID string) (models.App, error) {
	const op = "storage.pg.App"
	var app models.App

	err := s.Pool.QueryRow(ctx,
		"SELECT id, name, secret FROM users WHERE id = $1",
		appID).Scan(&app.ID, &app.Name, &app.Secret)

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

	err := s.Pool.QueryRow(ctx,
		"SELECT id, user_id, refresh_token FROM users_tokens WHERE refresh_token =$1",
		refreshToken).Scan(&token.ID, &token.UserID, &token.Token)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.RefreshToken{}, fmt.Errorf("%s: %w", op, storage.ErrRefreshTokenNotFound)
		}
		return models.RefreshToken{}, fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (s *Storage) SaveRefreshToken(ctx context.Context, refreshToken string, uid string) (models.RefreshToken, error) {
	const op = "storage.pg.SaveRefreshToken"
	var token models.RefreshToken

	err := s.Pool.QueryRow(ctx,
		"INSERT INTO users_tokens (user_id, refresh_token) VALUES ($1, $2) RETURNING id",
		uid, refreshToken).Scan(&token.ID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.RefreshToken{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		return models.RefreshToken{}, fmt.Errorf("%s: %w", op, err)
	}

	token.UserID = uid
	token.Token = refreshToken

	return token, nil
}

// IsUniqueViolation check is pgx error is about unique constraint violation
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

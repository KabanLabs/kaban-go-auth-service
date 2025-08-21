package pg

import (
	"context"

	"github.com/VACdotCS/kaban-go-auth-service/internal/domain/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, dbUrl string) *Storage {
	config, err := pgxpool.ParseConfig(dbUrl)

	if err != nil {
		panic(err)
	}

	pool, connErr := pgxpool.NewWithConfig(ctx, config)

	if connErr != nil {
		panic(connErr)
	}

	return &Storage{Pool: pool}
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (uid string, err error) {
	//TODO implement me
	panic("implement me")
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Storage) App(ctx context.Context, appID string) (models.App, error) {
	//TODO implement me
	panic("implement me")
}

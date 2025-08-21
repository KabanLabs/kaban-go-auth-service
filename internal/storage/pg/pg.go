package pg

import (
	"context"

	"github.com/VACdotCS/kaban-go-auth-service/internal/domain/models"
	"github.com/jackc/pgx/v5"
)

type Storage struct {
	conn *pgx.Conn
}

func New(ctx context.Context, dbUrl string) *Storage {
	config, err := pgx.ParseConfig(dbUrl)

	if err != nil {
		panic(err)
	}

	config.TLSConfig = nil

	conn, connErr := pgx.ConnectConfig(ctx, config)

	if connErr != nil {
		panic(connErr)
	}

	return &Storage{conn: conn}
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

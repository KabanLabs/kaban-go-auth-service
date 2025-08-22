package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	grpcapp "github.com/VACdotCS/kaban-go-auth-service/internal/app/grpc"
	"github.com/VACdotCS/kaban-go-auth-service/internal/config"
	"github.com/VACdotCS/kaban-go-auth-service/internal/services/auth"
	"github.com/VACdotCS/kaban-go-auth-service/internal/storage/pg"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	GRPCSrv *grpcapp.App
	DBPool  *pgxpool.Pool
}

func New(
	log *slog.Logger,
	grpcPort int,
	pgConfig config.PostgresConfig,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *App {

	dbUrl := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		pgConfig.User,
		pgConfig.Password,
		pgConfig.Host,
		pgConfig.Port,
		pgConfig.DbName,
	)

	storage, err := pg.New(context.Background(), dbUrl)

	if err != nil {
		panic(err)
	}

	authService := auth.New(
		log,
		storage,
		storage,
		storage,
		storage,
		storage,
		accessTokenTTL,
		refreshTokenTTL,
	)

	grpcApp := grpcapp.New(log, grpcPort, authService)

	return &App{
		GRPCSrv: grpcApp,
		DBPool:  storage.Pool,
	}
}

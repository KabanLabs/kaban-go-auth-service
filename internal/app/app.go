package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/VACdotCS/kaban-go-auth-service/internal/app/grpc"
	"github.com/VACdotCS/kaban-go-auth-service/internal/config"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	pgConfig config.PostgresConfig,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *App {
	// TODO: init db

	// TODO: init auth service (auth)

	grpcApp := grpcapp.New(log, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}

package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/VACdotCS/kaban-go-auth-service/internal/app"
	"github.com/VACdotCS/kaban-go-auth-service/internal/config"
	rsa_store "github.com/VACdotCS/kaban-go-auth-service/internal/lib/rsa-store"
)

const (
	envLocal = "local"
	envDev   = "envDev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)
	log.Info("Starting application")

	log.Info("Loading rsa key pair...")
	_, err := rsa_store.LoadOrGenerateKeys(2048, cfg.PrivateKeyTTL)

	if err != nil {
		log.Error("Failed to load rsa key pair", "error", err)
		panic(err)
	}

	application := app.New(
		log,
		cfg.GRPC.Port,
		cfg.PgConfig,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	go application.GRPCSrv.MustRun()

	// Graceful shutdown

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop

	log.Info("stopping application", slog.String("signal", sign.String()))

	application.GRPCSrv.Stop()
	application.DBPool.Close()

	log.Info("application stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

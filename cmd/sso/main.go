package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/VACdotCS/kaban-go-auth-service/internal/app"
	"github.com/VACdotCS/kaban-go-auth-service/internal/config"
	rsa_store "github.com/VACdotCS/kaban-go-auth-service/internal/lib/rsa-store"
)

const (
	envLocal = "local"
	envDev   = "envDev"
	envProd  = "prod"
)

func keysRotation(log *slog.Logger, ttl time.Duration) {
	const op = "main.keysRotation"

	logger := log.With(
		slog.String("op", op),
	)

	logger.Info("Starting key rotation worker")

	for {
		_, err := rsa_store.RotateKey(2048, ttl)

		if err == nil {
			logger.Info("Keys rotated successfully")
		}

		time.Sleep(time.Minute)
	}
}

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

	log.Info("Rsa key pair loaded")

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	application := app.New(
		ctx,
		cfg.Env,
		log,
		cfg.GRPC.Port,
		cfg.Http.Port,
		cfg.PgConfig,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
		cfg.Http.EnabledCors,
	)

	go application.GRPCSrv.MustRun()
	go application.HttpGateway.Run()
	go keysRotation(log, cfg.PrivateKeyTTL)

	// Graceful shutdown

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop

	log.Info("stopping application", slog.String("signal", sign.String()))

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	application.GRPCSrv.Stop(shutdownCtx)

	if err := application.HttpGateway.Stop(shutdownCtx); err != nil {
		log.Error("HTTP Gateway shutdown failed", "error", err)
	}

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

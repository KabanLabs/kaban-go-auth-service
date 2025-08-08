package main

import (
	"log/slog"
	"os"

	"github.com/VACdotCS/kaban-go-auth-service/internal/config"
)

const (
	envLocal = "local"
	envDev   = "envDev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()
	// TODO: logger init

	log := setupLogger(cfg.Env)
	log.Info("Starting application")

	// TODO: init app

	// TODO: run gRPC-server of the app
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

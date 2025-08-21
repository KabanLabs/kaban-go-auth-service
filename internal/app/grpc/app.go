package grpcapp

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/VACdotCS/kaban-go-auth-service/internal/config"
	authgrpc "github.com/VACdotCS/kaban-go-auth-service/internal/grpc/auth"
	"github.com/VACdotCS/kaban-go-auth-service/internal/services/auth"
	"github.com/VACdotCS/kaban-go-auth-service/internal/storage/pg"
	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

// Creates new gRPC server app
func New(
	log *slog.Logger,
	port int,
	pgConfig config.PostgresConfig,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *App {
	gRPCServer := grpc.NewServer()

	// Create db url from config
	dbUrl := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		pgConfig.User,
		pgConfig.Password,
		pgConfig.Host,
		pgConfig.Port,
		pgConfig.DbName,
	)

	fmt.Println(dbUrl)

	// TODO: where I should create context?
	storage := pg.New(context.Background(), dbUrl)

	authService := auth.New(log, storage, storage, storage, accessTokenTTL, refreshTokenTTL)

	authgrpc.Register(gRPCServer, authService)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("gRPC server is running", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).
		Info("stopping gRPC server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

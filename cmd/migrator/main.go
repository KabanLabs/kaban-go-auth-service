package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/VACdotCS/kaban-go-auth-service/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var migrationsPath, migrationType string

	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migrationType, "type", "", "up or down migration")
	flag.Parse()

	if migrationsPath == "" {
		panic("migrations-path is required")
	}

	if migrationType == "" {
		migrationType = "up"
	}

	cfg := config.MustLoad()

	var dbURL string

	if cfg.Env == "prod" {
		dbURL = fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=%s&sslrootcert=%s",
			cfg.PgConfig.User,
			cfg.PgConfig.Password,
			cfg.PgConfig.Host,
			cfg.PgConfig.Port,
			cfg.PgConfig.DbName,
			cfg.PgConfig.SSLMode,
			cfg.PgConfig.SSLRootCertPath,
		)
	} else {
		dbURL = fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			cfg.PgConfig.User,
			cfg.PgConfig.Password,
			cfg.PgConfig.Host,
			cfg.PgConfig.Port,
			cfg.PgConfig.DbName,
		)
	}

	m, err := migrate.New(
		"file://"+migrationsPath,
		dbURL,
	)
	if err != nil {
		log.Fatalf("Failed to create migrator: %v", err)
	}

	if err != nil {
		panic(err)
	}

	switch migrationType {
	case "up":
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("no migrations to apply")

				return
			}

			panic(err)
		}
	case "down":
		if err := m.Down(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("no migrations to apply")

				return
			}

			panic(err)
		}
	default:
		panic("unknown migration type")
	}

	fmt.Println("migrations applied")
}

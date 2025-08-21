package main

import (
	"errors"
	"flag"
	"fmt"

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

	m, err := migrate.New(
		"file://"+migrationsPath,
		fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			"sso_user",
			"sso_db_password",
			"localhost",
			5432,
			"sso",
		),
	)

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

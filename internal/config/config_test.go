package config

import (
	"os"
	"testing"
)

func TestMustLoad(t *testing.T) {
	tempFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	yamlContent := `
env: "test"
grpc:
  port: 44044
  timeout: 5s
pg_config:
  port: 5432
  host: "localhost"
  user: "user"
  password: "password"
  db_name: "test"
`
	if _, err := tempFile.Write([]byte(yamlContent)); err != nil {
		t.Fatal(err)
	}
	tempFile.Close()

	os.Setenv("CONFIG_PATH", tempFile.Name())
	defer os.Unsetenv("CONFIG_PATH")

	// Avoid command-line arguments interfering with the flag parsing
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd"}

	cfg := MustLoad()

	if cfg.Env != "test" {
		t.Errorf("expected env test, got %s", cfg.Env)
	}
	if cfg.GRPC.Port != 44044 {
		t.Errorf("expected grpc port 44044, got %d", cfg.GRPC.Port)
	}
	if cfg.PgConfig.Port != 5432 {
		t.Errorf("expected pg port 5432, got %d", cfg.PgConfig.Port)
	}
}

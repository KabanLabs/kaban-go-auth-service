package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env             string         `yaml:"env" env-default:"local"`
	GRPC            GRPCConfig     `yaml:"grpc"`
	AccessTokenTTL  time.Duration  `yaml:"access_token_ttl"`
	RefreshTokenTTL time.Duration  `yaml:"refresh_token_ttl"`
	PgConfig        PostgresConfig `yaml:"pg_config"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DbName   string `yaml:"db_name"`
	SSLMode  string `yaml:"ssl_mode"`
}

func MustLoad() *Config {
	path := fetchConfigPath()

	if path == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exists: " + path)
	}

	var config Config

	if err := cleanenv.ReadConfig(path, &config); err != nil {
		panic("failed to read config file: " + err.Error())
	}

	return &config
}

// fetchConfigPath fetches config path from cmd flag or env vars
func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}

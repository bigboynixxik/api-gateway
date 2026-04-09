package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	HTTPPort string `env:"HTTP_PORT" envDefault:"8080"`
	GRPCPort string `env:"GRPC_PORT" envDefault:"50051"`
	AppEnv   string `env:"APP_ENV" envDefault:"dev"`
}

func LoadConfig() (*Config, error) {
	var cfg Config

	err := env.Parse(&cfg)
	if err != nil {
		return nil, fmt.Errorf("config.LoadConfig failed to parse config: %w", err)
	}
	return &cfg, nil
}

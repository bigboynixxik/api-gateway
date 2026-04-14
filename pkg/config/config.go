package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort         string `env:"HTTP_PORT" envDefault:"8080"`
	AppEnv           string `env:"APP_ENV" envDefault:"dev"`
	EventServiceAddr string `env:"EVENT_SERVICE_ADDR" envDefault:"localhost:50051"`
	AuthServiceAddr  string `env:"AUTH_SERVICE_ADDR" envDefault:"localhost:50052"`
	JWTSecret        string `env:"JWT_SECRET"`
}

func LoadConfig(path string) (*Config, error) {
	_ = godotenv.Load(path)
	//if err != nil {
	//	return nil, fmt.Errorf("config.LoadConfig: %w", err)
	//}
	var cfg Config

	err := env.Parse(&cfg)
	if err != nil {
		return nil, fmt.Errorf("config.LoadConfig failed to parse config: %w", err)
	}
	return &cfg, nil
}

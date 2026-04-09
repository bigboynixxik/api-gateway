package app

import (
	"api-gateway/pkg/closer"
	"api-gateway/pkg/config"
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"google.golang.org/grpc"
)

type App struct {
	grpcServer *grpc.Server
	httpServer *http.Server
	grpcPort   string
	httpPort   string
	logs       *slog.Logger
	closer     *closer.Closer
}

func NewApp(ctx context.Context) (*App, error) {
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		return nil, err
	}
	fmt.Println(cfg.AppEnv, cfg.HTTPPort, cfg.GRPCPort)
	return nil, nil
}

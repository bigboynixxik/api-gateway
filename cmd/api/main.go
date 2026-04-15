package main

import (
	"context"

	"api-gateway/internal/app"
)

// @title Eventify API Gateway
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @version 1.0
// @description API для сервиса мероприятий Eventify
// @host localhost:8080
// @BasePath /
func main() {
	ctx := context.Background()
	a, err := app.NewApp(ctx)
	if err != nil {
		panic(err)
	}
	a.Run()
}

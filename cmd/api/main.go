package main

import (
	"api-gateway/internal/app"
	"context"
)

// @title Eventify API Gateway
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

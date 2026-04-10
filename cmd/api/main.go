package main

import (
	"api-gateway/internal/app"
	"context"
)

func main() {
	ctx := context.Background()
	a, err := app.NewApp(ctx)
	if err != nil {
		panic(err)
	}
	a.Run()
}

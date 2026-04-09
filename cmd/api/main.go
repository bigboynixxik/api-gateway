package main

import (
	"api-gateway/internal/app"
	"context"
)

func main() {
	ctx := context.Background()
	_, err := app.NewApp(ctx)
	if err != nil {
		panic(err)
	}

}

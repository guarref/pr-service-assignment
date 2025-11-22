package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/guarref/pr-service-assignment/config"
	"github.com/guarref/pr-service-assignment/internal/app"
)

func main() {
	
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg := config.MustLoad()

	application, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatalf("error creating new application: %v", err)
	}

	if err := application.Run(ctx); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

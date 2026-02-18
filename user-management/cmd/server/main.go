package main

import (
	"context"
	"flag"
	"os"

	"github.com/gofiber/fiber/v2/log"

	app "github.com/muazwzxv/user-management/internal"
)

func main() {
	mode := flag.String("mode", getEnvOrDefault("RUN_MODE", "http"), "Run mode: http, worker, or both")
	flag.Parse()

	runMode := app.RunMode(*mode)

	application, err := app.Init()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	log.Infow("starting application", "mode", runMode)

	if err := application.Start(context.Background(), runMode); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

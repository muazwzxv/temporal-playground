package main

import (
	"github.com/gofiber/fiber/v2/log"

	app "github.com/muazwzxv/user-management/internal"
)

func main() {
	application, err := app.Init()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	if err := application.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

package main

import (
	"log"

	"github.com/YubiApp/internal/config"
	"github.com/YubiApp/internal/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize and start the server
	srv := server.New(cfg)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 
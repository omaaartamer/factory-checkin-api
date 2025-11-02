package main

import (
	"log"

	"github.com/omaaartamer/factory-checkin-api/internal/repository"
	"github.com/omaaartamer/factory-checkin-api/pkg/config"
)

func main() {
	// Load configuration
	cfg := config.Load()

	log.Printf("Starting Factory Check-in API on port %s", cfg.Port)
	log.Printf("Connecting to database: %s", cfg.DatabaseURL)

	// Initialize repository (this will create tables)
	repo, err := repository.NewRepository(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repo.Close()

	log.Println("Database connection successful!")
	log.Println("Tables created/verified!")
	log.Printf("Application ready on port %s", cfg.Port)

	// Keep the app running
	select {}
}

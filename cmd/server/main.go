package main

import (
	"log"

	"github.com/omaaartamer/factory-checkin-api/internal/queue"
	"github.com/omaaartamer/factory-checkin-api/internal/repository"
	"github.com/omaaartamer/factory-checkin-api/internal/service"
	"github.com/omaaartamer/factory-checkin-api/pkg/config"
)

func main() {
	cfg := config.Load()
	log.Printf("Starting Factory Check-in API on port %s", cfg.Port)

	// Initialize repository
	repo, err := repository.NewRepository(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repo.Close()

	// Initialize queue
	q := queue.NewInMemoryQueue()
	defer q.Close()

	// Initialize service
	checkinService := service.NewCheckinService(repo, q)

	log.Println("Database connected!")
	log.Println("Queue initialized!")
	log.Println("Business logic ready!")

	// Test check-in
	response, err := checkinService.ProcessCheckin("EMP001")
	if err != nil {
		log.Printf("Test failed: %v", err)
	} else {
		log.Printf("Test passed: %s", response.Message)
	}

	log.Printf("Application ready!")
	select {}
}

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/omaaartamer/factory-checkin-api/internal/handler"
	"github.com/omaaartamer/factory-checkin-api/internal/queue"
	"github.com/omaaartamer/factory-checkin-api/internal/repository"
	"github.com/omaaartamer/factory-checkin-api/internal/service"
	"github.com/omaaartamer/factory-checkin-api/internal/worker"
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
	// q := queue.NewInMemoryQueue()
	// defer q.Close()
	// Initialize Redis queue
	q, err := queue.NewRedisQueue(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to initialize Redis queue: %v", err)
	}
	defer q.Close()

	// Initialize service
	checkinService := service.NewCheckinService(repo, q)

	// Initialize background worker
	bgWorker := worker.NewWorker(q, cfg)
	bgWorker.Start()
	defer bgWorker.Stop()

	// Initialize HTTP handler
	h := handler.NewHandler(checkinService)
	router := h.SetupRoutes()

	log.Println("Database Connected!")
	log.Println("Redis Queue Connected!")
	log.Println("Business logic ready!")

	// // Test check-in
	// response, err := checkinService.ProcessCheckin("EMP001")
	// if err != nil {
	// 	log.Printf("Test failed: %v", err)
	// } else {
	// 	log.Printf("Test passed: %s", response.Message)
	// }
	// Start HTTP server

	// Graceful shutdown
	go func() {
		if err := router.Run(":" + cfg.Port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
}

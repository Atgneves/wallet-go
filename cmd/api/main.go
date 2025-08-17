package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wallet-go/internal/operation"
	"wallet-go/internal/router"
	"wallet-go/internal/shared/config"
	"wallet-go/internal/shared/database"
	"wallet-go/internal/shared/kafka"
	"wallet-go/internal/shared/utils"
	"wallet-go/internal/wallet"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize MongoDB
	mongoClient, err := database.NewMongoClient(cfg.MongoDB.URI)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer mongoClient.Disconnect(context.Background())

	// Initialize stores
	walletStore := wallet.NewStore(mongoClient)
	operationStore := operation.NewStore(mongoClient)

	// Initialize validator and lock manager
	walletValidator := wallet.NewValidator()
	lockManager := utils.NewWalletLockManager()

	// Initialize wallet service
	walletService := wallet.NewService(walletStore, operationStore, walletValidator, lockManager)

	// Create service adapter for kafka
	walletServiceAdapter := wallet.NewServiceAdapter(walletService)

	// Initialize Kafka
	kafkaProducer, err := kafka.NewProducer(cfg.Kafka.Brokers)
	if err != nil {
		log.Fatal("Failed to create Kafka producer:", err)
	}
	defer kafkaProducer.Close()

	// Criar tópicos automaticamente
	topics := []string{
		cfg.Kafka.Topics.Deposit,
		cfg.Kafka.Topics.Withdraw,
		cfg.Kafka.Topics.Transfer,
	}

	if err := kafka.CreateTopics(cfg.Kafka.Brokers, topics); err != nil {
		log.Printf("Warning: Could not create Kafka topics: %v", err)
	} else {
		log.Println("Kafka topics created/verified successfully")
	}

	kafkaConsumer, err := kafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID)
	if err != nil {
		log.Fatal("Failed to create Kafka consumer:", err)
	}
	defer kafkaConsumer.Close()

	// Set wallet service adapter in kafka consumer
	kafkaConsumer.SetWalletService(walletServiceAdapter)

	// Start Kafka consumers
	go kafkaConsumer.StartConsumers()

	// Setup router
	r := router.Setup(mongoClient, kafkaProducer, cfg)

	// Setup server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	log.Printf("Server started on port %s", cfg.Server.Port)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}

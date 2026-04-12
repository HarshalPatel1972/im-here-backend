package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/guardian/im-here/internal/config"
	"github.com/guardian/im-here/internal/db"
	"github.com/guardian/im-here/internal/poller"
	"github.com/joho/godotenv"
)

func main() {
	// Try loading .env for development
	_ = godotenv.Load()

	// Load config (panics if required vars missing)
	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Received shutdown signal, shutting down gracefully...")
		cancel()
	}()

	log.Printf("Connecting to database at %s...", cfg.DatabaseURL)
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer pool.Close()

	log.Println("I'm Here is running")

	p := poller.New(cfg, pool)
	go p.Start(ctx)

	// Wait for context cancellation
	<-ctx.Done()
	log.Println("Shutdown complete")
}

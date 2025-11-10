package main

import (
	"log"

	"github.com/mihir/transactions/internal/server"
	"github.com/mihir/transactions/pkg/config"
)

func main() {
	config := config.LoadConfig()

	if config.PostgresURL == "" {
		log.Fatal("POSTGRES_URL environment variable is required")
	}

	ctx := context.Background()
	dbPool, err := database.Connect(ctx, config.PostgresURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	if err := database.Migrate(ctx, dbPool); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	addr := ":" + config.Port
	if err := server.Start(addr, dbPool); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

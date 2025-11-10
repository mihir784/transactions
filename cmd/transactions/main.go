package main

import (
	"log"

	"github.com/mihir/transactions/internal/server"
	"github.com/mihir/transactions/pkg/config"
)

func main() {
	addr := ":" + config.GetPort()
	if err := server.Start(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

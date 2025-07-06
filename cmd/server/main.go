package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"unified-ads-mcp/internal/config"
	"unified-ads-mcp/internal/mcp"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	server := mcp.NewServer(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down server...")
		cancel()
	}()

	if err := server.Start(ctx); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

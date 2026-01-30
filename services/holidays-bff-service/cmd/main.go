package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"holidays-bff-service/config"
	"holidays-bff-service/router"
	observability "holidays-observability"
)

func main() {
	// Initialize OpenTelemetry
	ctx := context.Background()
	shutdown := observability.InitProvider(ctx, "holidays-bff-service")
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Printf("failed to shutdown OTEL: %v", err)
		}
	}()

	// Load configuration (allow override via CONFIG_PATH)
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create router
	r := router.NewRouter(cfg)
	r.Setup()

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Starting BFF service on %s", addr)
	if err := r.Engine().Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

package main

import (
	"fmt"
	"log"
	"os"

	"holidays-bff-service/config"
	"holidays-bff-service/router"
)

func main() {
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

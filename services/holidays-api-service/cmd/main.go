package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"holidays-api-service/config"
	"holidays-api-service/controller"
	"holidays-api-service/repository"
	"holidays-api-service/router"
	"holidays-api-service/service"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load configuration (allow override via CONFIG_PATH)
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// MongoDB client
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoDB.URI))
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	defer func() {
		_ = client.Disconnect(context.Background())
	}()

	db := client.Database(cfg.MongoDB.Database)

	// Layers
	repo := repository.NewHolidayRepository(db, cfg.MongoDB.Collection)
	svc := service.NewHolidayService(repo, cfg.ExternalAPI.URL)
	ctrl := controller.NewHolidayController(svc)

	r := router.NewRouter(cfg)
	r.Setup(ctrl)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("starting holidays-api-service on %s", addr)
	if err := r.Engine().Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

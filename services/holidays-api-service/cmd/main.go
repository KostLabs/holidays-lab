package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"holidays-api-service/config"
	"holidays-api-service/controller"
	"holidays-api-service/repository"
	"holidays-api-service/router"
	"holidays-api-service/service"
	observability "holidays-observability"

	"github.com/KostLabs/golog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
)

func main() {
	// Initialize OpenTelemetry
	ctx := context.Background()
	shutdown := observability.InitProvider(ctx, "holidays-api-service")
	defer func() {
		if err := shutdown(ctx); err != nil {
			golog.Error("failed to shutdown OTEL", map[string]any{"error": err.Error()})
		}
	}()

	// Load configuration (allow override via CONFIG_PATH)
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		golog.Error("failed to load config", map[string]any{"error": err.Error()})
		os.Exit(1)
	}

	// MongoDB client with OTEL monitoring
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(cfg.MongoDB.URI)
	clientOpts.Monitor = otelmongo.NewMonitor()

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		golog.Error("failed to connect to MongoDB", map[string]any{"error": err.Error()})
		os.Exit(1)
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
	golog.Info("starting holidays-api-service", map[string]any{"addr": addr})
	if err := r.Engine().Run(addr); err != nil {
		golog.Error("server failed", map[string]any{"error": err.Error()})
		os.Exit(1)
	}
}

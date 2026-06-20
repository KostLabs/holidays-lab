package main

import (
	"context"
	"fmt"
	"os"

	"holidays-bff-service/config"
	"holidays-bff-service/router"
	observability "holidays-observability"

	"github.com/KostLabs/golog"
)

func main() {
	ctx := context.Background()
	shutdown := observability.InitProvider(ctx, "holidays-bff-service")
	defer func() {
		shutdownErr := shutdown(ctx)
		if shutdownErr != nil {
			golog.Error("failed to shutdown OTEL", map[string]any{"error": shutdownErr.Error()})
		}
	}()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		golog.Error("failed to load config", map[string]any{"error": err.Error()})
		os.Exit(1)
	}

	r := router.NewRouter(cfg)
	r.Setup()

	addr := fmt.Sprintf(":%d", cfg.Port)
	golog.Info("starting holidays-bff-service", map[string]any{"addr": addr})
	runErr := r.Engine().Run(addr)
	if runErr != nil {
		golog.Error("server failed", map[string]any{"error": runErr.Error()})
		os.Exit(1)
	}
}

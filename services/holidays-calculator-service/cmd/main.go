package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"holidays-calculator-service/config"
	"holidays-calculator-service/controller"
	"holidays-calculator-service/pkg/holidaysclient"
	"holidays-calculator-service/router"
	"holidays-calculator-service/service"
	observability "holidays-observability"
)

func main() {
	// Initialize OpenTelemetry
	ctx := context.Background()
	shutdown := observability.InitProvider(ctx, "holidays-calculator-service")
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
		log.Fatalf("failed to load config: %v", err)
	}

	holidaysClient := holidaysclient.NewClient(cfg.HolidaysAPIService.URL)
	svc := service.NewCalculatorService(holidaysClient)
	ctrl := controller.NewCalculatorController(svc)
	r := router.NewRouter(cfg)
	r.Setup(ctrl)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("starting holidays-calculator-service on %s", addr)
	if err := r.Engine().Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

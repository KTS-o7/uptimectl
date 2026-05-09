package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	configPath := "/opt/uptimectl/config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Printf("uptimectl starting — monitoring %d services every %s", len(cfg.Services), cfg.Interval)
	for _, svc := range cfg.Services {
		log.Printf("  → %s: %s (%s)", svc.Name, svc.URL, svc.Method)
	}

	store, err := NewResultsStore(cfg.DataPath, time.Duration(cfg.HistoryDays)*24*time.Hour)
	if err != nil {
		log.Fatalf("failed to create store: %v", err)
	}

	// Signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Run initial check immediately
	runChecks(cfg, store)

	// Then run on the configured interval
	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			runChecks(cfg, store)
		case sig := <-sigCh:
			log.Printf("received signal %v, shutting down", sig)
			return
		}
	}
}

func runChecks(cfg *Config, store *ResultsStore) {
	log.Printf("running health checks...")

	for _, svc := range cfg.Services {
		result := CheckService(svc)
		status := "UP"
		if !result.Success {
			status = "DOWN"
		}
		log.Printf("  %s: %s (%dms)", result.ServiceName, status, result.LatencyMs)
		if result.Error != "" {
			log.Printf("    └─ %s", result.Error)
		}
		store.AddResult(result)
	}

	// Generate status page
	if err := GenerateStatusPage(store, cfg.Services, cfg.OutputPath); err != nil {
		log.Printf("error generating status page: %v", err)
	} else {
		log.Printf("status page written to %s", cfg.OutputPath)
	}
}

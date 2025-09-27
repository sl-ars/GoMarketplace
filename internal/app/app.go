package app

import (
	"go-app-marketplace/internal/app/config"
	"go-app-marketplace/internal/app/connections"
	"go-app-marketplace/internal/app/start"
	"log"
)

func Run(configFiles ...string) {
	//ctx := context.Background()

	cfg, err := config.NewConfig(configFiles...)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	conns, err := connections.NewConnections(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize connections: %v", err)
	}
	defer conns.Close()

	start.StartHTTPServer(cfg.HTTPServer.Port)
}

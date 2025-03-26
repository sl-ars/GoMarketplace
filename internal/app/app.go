package app

import (
	"go-app-marketplace/internal/app/config"
	"go-app-marketplace/internal/app/connections"
	"go-app-marketplace/internal/app/start"
	"go-app-marketplace/internal/deliveries/http"
	"go-app-marketplace/internal/repositories"
	"go-app-marketplace/internal/services"
	"go-app-marketplace/internal/usecases"
	"log"
)

func Run(configFiles ...string) {
	// Load env
	cfg, err := config.NewConfig(configFiles...)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize DB and external connections
	conns, err := connections.NewConnections(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize connections: %v", err)
	}
	defer conns.Close()

	// DB migrations
	//if err := connections.RunMigrations(cfg.DB.DSN); err != nil {
	//	log.Fatalf("Failed to run migrations: %v", err)
	//}

	// Dependency injection
	userRepo := repositories.NewUserPostgresRepo(conns.DB)
	userUC := usecases.NewUserUseCase(userRepo)
	userService := services.NewUserService(userUC, cfg.JWTSecret)

	// Wrap services
	svc := &http.Services{
		User: userService,
	}

	// Router
	router := http.NewRouter(svc)

	// Start the server
	start.StartHTTPServer(cfg.HTTPServer.Port, router)
}

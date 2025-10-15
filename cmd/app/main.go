package main

import (
	"flag"
	_ "go-app-marketplace/docs"
	"go-app-marketplace/internal/app"
)

// @title Go Marketplace API
// @version 1.0
// @description API documentation for Marketplace project
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	configFile := flag.String("config", "./configs/.env", "Path to configuration file")
	flag.Parse()

	app.Run(*configFile)
}

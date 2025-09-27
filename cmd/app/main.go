package main

import (
	"flag"
	_ "go-app-marketplace/docs"
	"go-app-marketplace/internal/app"
)

// @title Online Marketplace
// @version 0.1
// @description This is the API documentation for the online marketplace.
// @host localhost:8080
// @BasePath /
func main() {
	configFile := flag.String("config", "./configs/.env", "Path to configuration file")
	flag.Parse()

	app.Run(*configFile)
}

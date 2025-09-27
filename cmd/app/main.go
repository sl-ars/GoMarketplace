package main

import (
	"flag"
	"go-app-marketplace/internal/app"
)

func main() {
	configFile := flag.String("config", "./configs/.env", "Path to configuration file")
	flag.Parse()

	app.Run(*configFile)
}

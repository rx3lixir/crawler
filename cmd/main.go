package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/rx3lixir/crawler/appconfig"
	"github.com/rx3lixir/crawler/telegram"
)

func main() {
	configFile := flag.String("config", "", "Path to config file")
	flag.Parse()

	// Loading data
	if err := appconfig.LoadConfig(*configFile); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Println("Configuration loaded successfully")

	// Running bot instance
	telegram.StartBot(*appconfig.CrawlerApp)
}

package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/rx3lixir/crawler/appconfig"
	"github.com/rx3lixir/crawler/telegram"
)

func main() {
	// Getting tg toke and other stuff from args
	telegramToken := flag.String("token", "", "Telegram bot token")
	flag.Parse()

	if *telegramToken == "" {
		log.Fatalf("Telegram Bot token is required")
	}

	// Log for proper inusrence
	fmt.Println("Telegram token recieved:", *telegramToken)

	// Loading data and providing telegramToken
	if err := appconfig.LoadConfig(*telegramToken); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Println("Configuration loaded successfully")

	// Running bot instance
	telegram.StartBot(*appconfig.CrawlerApp)
}

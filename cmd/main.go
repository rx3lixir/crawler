package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/rx3lixir/crawler/telegram"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	telegramToken := os.Getenv("TG_BOT_TOKEN")
	telegram.StartBot(telegramToken)
}

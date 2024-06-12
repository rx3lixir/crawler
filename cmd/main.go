package main

import (
	"log"

	"github.com/rx3lixir/crawler/appconfig"
	"github.com/rx3lixir/crawler/telegram"
)

func main() {
	// Загружаем конфиг приложения и все .env файлы
	if err := appconfig.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	telegram.StartBot(*appconfig.CrawlerApp)
}

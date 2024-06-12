package appconfig

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	TelegramToken string
	GoogleAuthKey string
	SpreadsheetID string
}

// Создаем инстанс для приложения
var CrawlerApp *AppConfig

// Загружает конфигурации и .env файлы
func LoadConfig() error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("Error loading .env file: %v", err)
	}

	CrawlerApp = &AppConfig{
		TelegramToken: os.Getenv("TG_BOT_TOKEN"),
		GoogleAuthKey: os.Getenv("GOOGLE_AUTH_KEY"),
		SpreadsheetID: os.Getenv("SPREADSHEET_ID"),
	}

	if CrawlerApp.TelegramToken == "" || CrawlerApp.GoogleAuthKey == "" || CrawlerApp.SpreadsheetID == "" {
		return fmt.Errorf("incomplete configuration: missing environment variables")
	}

	return nil
}

type SiteConfig struct {
	UrlToVisit        string
	EventType         string
	AnchestorSelector string
	TitleSelector     string
	DateSelector      string
	LocationSelector  string
	LinkSelector      string
}

type EventConfig struct {
	Title     string `json:"title"`
	Date      string `json:"date"`
	Location  string `json:"location"`
	Link      string `json:"link"`
	EventType string `json:"eventType"`
}

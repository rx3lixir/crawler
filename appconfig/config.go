package appconfig

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	TelegramToken string
	GoogleAuthKey string
	SpreadsheetID string
}

// Создаем инстанс для приложения
var AppInstance *AppConfig

// Загружает конфигурации и .env файлы
func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	AppInstance = &AppConfig{
		TelegramToken: os.Getenv("TG_BOT_TOKEN"),
		GoogleAuthKey: os.Getenv("GOOGLE_AUTH_KEY"),
		SpreadsheetID: os.Getenv("SPREADSHEET_ID"),
	}
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

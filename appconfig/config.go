package appconfig

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

type AppConfig struct {
	TelegramToken string `json:"telegram_token"`
	GoogleAuthKey string `json:"google_auth_key"`
	SpreadsheetID string `json:"spreadsheet_id"`
}

var CrawlerApp *AppConfig

func LoadConfig(configPath string) error {
	if configPath != "" {
		return loadConfigFromFile(configPath)
	}
	return loadConfigFromEnv()
}

func loadConfigFromFile(configPath string) error {
	file, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("Error reading config file: %v", err)
	}

	config := &AppConfig{}
	if err := json.Unmarshal(file, config); err != nil {
		return fmt.Errorf("Error parsing config file: %v", err)
	}

	CrawlerApp = config
	return validateConfig()
}

func loadConfigFromEnv() error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("Error loading .env file: %v", err)
	}

	CrawlerApp = &AppConfig{
		TelegramToken: os.Getenv("TELEGRAM_TOKEN"),
		GoogleAuthKey: os.Getenv("GOOGLE_AUTH_KEY"),
		SpreadsheetID: os.Getenv("SPREADSHEET_ID"),
	}

	return validateConfig()
}

func validateConfig() error {
	if CrawlerApp.TelegramToken == "" || CrawlerApp.GoogleAuthKey == "" || CrawlerApp.SpreadsheetID == "" {
		return fmt.Errorf("incomplete configuration: missing required values")
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

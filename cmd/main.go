package main

import (
	"github.com/rx3lixir/crawler/appconfig"
	"github.com/rx3lixir/crawler/telegram"
)

func main() {
	// Загружаем конфиг приложения и все .env файлы
	appconfig.LoadConfig()

	appInstance := *appconfig.AppInstance

	telegram.StartBot(appInstance.TelegramToken)
}

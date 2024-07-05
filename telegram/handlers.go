package telegram

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/rx3lixir/crawler/appconfig"
	"github.com/rx3lixir/crawler/spreadsheets"
	"github.com/rx3lixir/crawler/web"
)

// Переменная для хранения пользовательских конфигураций для поиска
var userConfigs []appconfig.SiteConfig

// Отправляет пользователю сообщение в tg
func sendMessageHandler(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

// Сбрасывает конфигурационный файл
func resetConfigHandler() {
	// Обновляем глобальную переменную userConfigs
	userConfigs = nil
}

// Запускает веб-скраппинг применяя конфигурацию
func runWebScraperHandler(bot *tgbotapi.BotAPI, chatID int64, crawlerAppConfig appconfig.AppConfig) {
	var siteConfigs []appconfig.SiteConfig

	if len(userConfigs) > 0 {
		siteConfigs = userConfigs
	} else {
		sendMessageHandler(bot, chatID, "Конфигурация не задана. Запустите /config чтобы добавить файл конфигурации")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	allEvents := web.WebScraper(crawlerAppConfig, ctx, siteConfigs)
	spreadsheets.WriteToSpreadsheet(allEvents, *&crawlerAppConfig)

	sendMessageHandler(bot, chatID, "Ищейкин сделал дело. Проверьте результат по ссылке: https://docs.google.com/spreadsheets/d/1G8eLUjCeqBZ9dqQJiWxJ3GfjBS9Oqd4_lLnaRMsCbYo/edit#gid=0")
}

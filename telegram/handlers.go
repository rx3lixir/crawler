package telegram

import (
	"fmt"
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
func resetConfigHandler(bot *tgbotapi.BotAPI, chatID int64) {
	// Устанавливаем дефолтные конфигурации для поиска
	afishaConfigConcert := appconfig.SiteConfig{
		UrlToVisit:        "https://bar.afishagoroda.ru/events/festivali",
		EventType:         "Фестивали",
		AnchestorSelector: "div.events-elem",
		TitleSelector:     "a.title",
		DateSelector:      "div.date",
		LocationSelector:  "div.place",
		LinkSelector:      "a.img-wrap",
	}

	afishaConfigTheater := appconfig.SiteConfig{
		UrlToVisit:        "https://bar.afishagoroda.ru/events/detiam",
		EventType:         "Детям",
		AnchestorSelector: "div.events-elem",
		TitleSelector:     "a.title",
		DateSelector:      "div.date",
		LocationSelector:  "div.place",
		LinkSelector:      "a.img-wrap",
	}

	// Обновляем глобальную переменную userConfigs
	userConfigs = nil

	// Пример использования переменных
	fmt.Printf("Конфиг сброшен, ищем: %s", afishaConfigConcert)
	fmt.Printf("Конфиг сброшен, ищем: %s", afishaConfigTheater)
}

// Запускает веб-скраппинг применяя конфигурацию
func runWebScraperHandler(bot *tgbotapi.BotAPI, chatID int64, crawlerAppConfig appconfig.AppConfig) {
	defaultConfigs := []appconfig.SiteConfig{
		{
			UrlToVisit:        "https://stonemir.ru/blog/",
			EventType:         "Клубы",
			AnchestorSelector: "div.article-inner",
			TitleSelector:     "h3.post",
			DateSelector:      "div.js-store-prod-descr",
			LocationSelector:  "div.js-product-price",
			LinkSelector:      "a",
		},
	}

	var siteConfigs []appconfig.SiteConfig

	if len(userConfigs) > 0 {
		siteConfigs = userConfigs
	} else {
		siteConfigs = defaultConfigs
	}

	allEvents := web.WebScraper(siteConfigs)
	spreadsheets.WriteToSpreadsheet(allEvents, *&crawlerAppConfig)

	sendMessageHandler(bot, chatID, "Ищейкин сделал дело. Проверьте результат по ссылке: https://docs.google.com/spreadsheets/d/1G8eLUjCeqBZ9dqQJiWxJ3GfjBS9Oqd4_lLnaRMsCbYo/edit#gid=0")
}

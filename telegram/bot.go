package telegram

import (
	"log"

	"github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/rx3lixir/crawler/config"
	"github.com/rx3lixir/crawler/spreadsheets"
	"github.com/rx3lixir/crawler/web"
)

func StartBot(token string) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Error initializing bot entity: %v", err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	updates, err := bot.GetUpdatesChan(tgbotapi.NewUpdate(0))
	if err != nil {
		log.Fatalf("Error getting updates: %v", err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		handleCommands(bot, update)
	}
}

func handleCommands(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	command := update.Message.Command()

	switch command {
	case "start":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Добро пожаловать! Запустите /run для запуска ищейкина")
		bot.Send(msg)
	case "run":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Запускаю веб-скраппинг...")
		bot.Send(msg)
		runWebScraper(bot, update.Message.Chat.ID)
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Хуйню ввели, отправьте /start для получения помощи")
		bot.Send(msg)
	}
}

func runWebScraper(bot *tgbotapi.BotAPI, chatID int64) {
	afishaConfigConcert := configs.SiteConfig{
		UrlToVisit:        "https://bar.afishagoroda.ru/events/koncert",
		EventType:         "Концерт",
		AnchestorSelector: "div.events-elem",
		TitleSelector:     "a.title",
		DateSelector:      "div.date",
		LocationSelector:  "div.place",
		LinkSelector:      "a.img-wrap",
	}

	afishaConfigTheater := configs.SiteConfig{
		UrlToVisit:        "https://bar.afishagoroda.ru/events/teatr",
		EventType:         "Театр",
		AnchestorSelector: "div.events-elem",
		TitleSelector:     "a.title",
		DateSelector:      "div.date",
		LocationSelector:  "div.place",
		LinkSelector:      "a.img-wrap",
	}

	SearchSitesConfigs := []configs.SiteConfig{afishaConfigConcert, afishaConfigTheater}

	allEvents := web.WebScraper(SearchSitesConfigs)
	spreadsheets.SaveDataToSpreadSheet(allEvents)

	msg := tgbotapi.NewMessage(chatID, "Ищейкин сделал дело")
	bot.Send(msg)
}

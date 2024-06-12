package telegram

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rx3lixir/crawler/appconfig"
)

func StartBot(crawlerAppConfig appconfig.AppConfig) {

	bot, err := tgbotapi.NewBotAPI(crawlerAppConfig.TelegramToken)
	if err != nil {
		log.Fatalf("Error initializing bot entity: %v", err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Счетчик для ожидания апдейта
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Обрабатываем апдейты приходящие из телеграма
	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("Received update: %v", update)

		handleCommands(bot, &update, crawlerAppConfig)
	}
}

func handleCommands(bot *tgbotapi.BotAPI, update *tgbotapi.Update, crawlerAppConfig appconfig.AppConfig) {
	// Переменная для хранения введенной пользователем команды
	command := strings.TrimSpace(strings.ToLower(update.Message.Command()))
	log.Printf("Received command: %s", command)

	// Обработка отправленного файла
	if update.Message.Document != nil {
		log.Println("Document received")
		handleFileUpload(bot, update)
		return
	}

	chatId := update.Message.Chat.ID

	switch command {
	case "start":
		sendMessageHandler(bot, chatId, "Добро Пожаловать! Запустите /run чтобы начать работу")
	case "run":
		sendMessageHandler(bot, chatId, "Запускаю веб-скраппинг!")
		runWebScraperHandler(bot, update.Message.Chat.ID, crawlerAppConfig)
	case "config":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, отправьте файл с конфигурациями")
		msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true, Selective: true}
		bot.Send(msg)
	case "reset":
		resetConfigHandler(bot, update.Message.Chat.ID)
	default:
		sendMessageHandler(bot, chatId, "Что-то пошло не так... Может не верно ввели команду?")
	}
}

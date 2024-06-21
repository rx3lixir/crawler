package telegram

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rx3lixir/crawler/appconfig"
	"github.com/rx3lixir/crawler/spreadsheets"
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
		sendMessageHandler(bot, chatId, "Добро Пожаловать! Перед началом рекомендую сбросить текущие конфигурации с помощью /reset и очистить таблицу с помощью /clear. Запустите /config чтобы задать конфигурацию и /run чтобы запустить поиск!")
	case "run":
		sendMessageHandler(bot, chatId, "Запускаю веб-скраппинг! Пожалуйста подождите, обычно это занимает не более 2 минут")
		runWebScraperHandler(bot, update.Message.Chat.ID, crawlerAppConfig)
	case "config":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, отправьте файл с конфигурациями. Проверьте чтобы он файл был в формате .json")
		msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true, Selective: true}
		bot.Send(msg)
	case "reset":
		resetConfigHandler()
		sendMessageHandler(bot, chatId, "Конфигурации успешно сброшены! Нажмите /config чтобы задать новые!")
	case "clear":
		sendMessageHandler(bot, chatId, "Очищаю таблицу...")
		spreadsheets.ClearAllSheets(crawlerAppConfig)
		sendMessageHandler(bot, chatId, "Листы в таблице очищены! Пора что-нибудь найти и скорее их заполнить!")
	default:
		sendMessageHandler(bot, chatId, "Что-то пошло не так... Может не верно ввели команду?")
	}
}

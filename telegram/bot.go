package telegram

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

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
	case "config":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, отправьте файл с конфигурациями")
		msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true, Selective: true}

		sentMsg, err := bot.Send(msg)
		if err != nil {
			log.Printf("Error sending message: %v", err)
			return
		}

		clearKeyboardMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		clearKeyboardMsg.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{
			RemoveKeyboard: true,
			Selective:      true,
		}
		bot.Send(clearKeyboardMsg)

		updates := bot.ListenForWebhook("/" + bot.Token)
		go http.ListenAndServe("0.0.0.0:8080", nil)

		for update := range updates {
			if update.Message == nil || update.Message.ReplyToMessage.MessageID != sentMsg.MessageID {
				continue
			}

			if update.Message.Document != nil {
				handleConfigFileUpload(bot, update)
				break
			}
		}
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Хуйню ввели, отправьте /start для получения помощи")
		bot.Send(msg)
	}
}

// Переменная для хранения пользовательских конфигураций для поиска
var userConfigs []configs.SiteConfig

func handleConfigFileUpload(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	// Получаем файл, отправленный пользователем
	if update.Message.Document == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы не прикрепили файл с конфигурациями")
		bot.Send(msg)
		return
	}

	fileID := update.Message.Document.FileID

	fileInfo, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка получения файла")
		bot.Send(msg)
		return
	}

	// Проверяем расширение файла
	if !strings.HasSuffix(fileInfo.FilePath, ".json") {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Файл должен иметь расширение .json")
		bot.Send(msg)
		return
	}

	fileURL, err := bot.GetFileDirectURL(fileInfo.FilePath)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка получения URL файла")
		bot.Send(msg)
		return
	}

	// Скачиваем файл с конфигурациями
	resp, err := http.Get(fileURL)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка скачивания файла")
		bot.Send(msg)
		return
	}
	defer resp.Body.Close()

	// Разбираем JSON-файл на структуры configs.SiteConfig
	var siteConfigs []configs.SiteConfig
	err = json.NewDecoder(resp.Body).Decode(&siteConfigs)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка разбора JSON")
		bot.Send(msg)
		return
	}

	// Сохраняем конфигурации для использования в runWebScraper
	userConfigs = siteConfigs
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Конфигурации успешно загружены")
	bot.Send(msg)
}

func runWebScraper(bot *tgbotapi.BotAPI, chatID int64) {
	// Дефолтная конфигурация для поиска
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

	// Создаем переменную для хранения конфигураций сайтов предназначенных для поиска
	var siteConfigs []configs.SiteConfig

	// Используем конфигурации, загруженные пользователем, если они есть
	if len(userConfigs) > 0 {
		siteConfigs = userConfigs
	} else {
		// Если пользовательские конфигурации отсутствуют, используем конфигурации по умолчанию
		siteConfigs = []configs.SiteConfig{afishaConfigConcert, afishaConfigTheater}
	}

	// Ищем ивенты по конфигурации
	allEvents := web.WebScraper(siteConfigs)

	// Cохраняем все что нашли в google таблицы
	spreadsheets.SaveDataToSpreadSheet(allEvents)

	msg := tgbotapi.NewMessage(chatID, "Ищейкин сделал дело. Проверьте результат по ссылке: https://docs.google.com/spreadsheets/d/1G8eLUjCeqBZ9dqQJiWxJ3GfjBS9Oqd4_lLnaRMsCbYo/edit#gid=0")
	bot.Send(msg)
}

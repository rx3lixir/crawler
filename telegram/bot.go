package telegram

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

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

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("Received update: %v", update)

		handleCommands(bot, update)
	}
}

func handleCommands(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	// Добавляем лог для проверки содержимого сообщения
	log.Printf("Message text: %s", update.Message.Text)

	command := strings.TrimSpace(strings.ToLower(update.Message.Command()))
	log.Printf("Received command: %s", command)

	// Обработка отправленного файла
	if update.Message.Document != nil {
		log.Println("Document received")
		handleConfigFileUpload(bot, update)
		return
	}

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
		bot.Send(msg)
		log.Println("Config command processed")
	case "reset": // Новая команда "resetConfig"
		resetConfig(bot, update.Message.Chat.ID)
	default:
		log.Printf("Unknown command: %s", command)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Повторите попытку")
		bot.Send(msg)
	}
}

// Переменная для хранения пользовательских конфигураций для поиска
var userConfigs []configs.SiteConfig

func handleConfigFileUpload(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	log.Println("Handling file upload")

	// Проверка, что это ответ на сообщение с запросом конфигурации
	if update.Message.ReplyToMessage != nil {
		log.Printf("Reply to message: %v", update.Message.ReplyToMessage)
	}

	// Получаем файл, отправленный пользователем
	if update.Message.Document == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы не прикрепили файл с конфигурациями")
		bot.Send(msg)
		return
	}

	fileID := update.Message.Document.FileID
	log.Printf("File ID: %s", fileID)

	fileInfo, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		log.Printf("Error getting file info: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка получения файла")
		bot.Send(msg)
		return
	}

	log.Printf("File info: %v", fileInfo)

	// Проверяем расширение файла
	if !strings.HasSuffix(fileInfo.FilePath, ".json") {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Файл должен иметь расширение .json")
		bot.Send(msg)
		return
	}

	log.Println("Getting file")

	// Скачиваем файл напрямую через Telegram API
	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", bot.Token, fileInfo.FilePath)
	log.Printf("File URL: %s", fileURL)

	resp, err := http.Get(fileURL)
	if err != nil {
		log.Printf("Error downloading file: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка скачивания файла")
		bot.Send(msg)
		return
	}
	defer resp.Body.Close()

	log.Println("File successfully downloaded")

	// Читаем содержимое файла
	fileBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading file: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка чтения файла")
		bot.Send(msg)
		return
	}

	// Разбираем JSON-файл на структуры configs.SiteConfig
	err = json.Unmarshal(fileBytes, &userConfigs)
	if err != nil {
		log.Printf("Error unmarshalling JSON: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка разбора JSON")
		bot.Send(msg)
		return
	}

	log.Println("Configurations successfully loaded")

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Конфигурации успешно загружены")
	bot.Send(msg)
}

func resetConfig(bot *tgbotapi.BotAPI, chatID int64) {
	// Устанавливаем дефолтные конфигурации для поиска

	afishaConfigConcert := configs.SiteConfig{
		UrlToVisit:        "https://bar.afishagoroda.ru/events/festivali",
		EventType:         "Фестивали",
		AnchestorSelector: "div.events-elem",
		TitleSelector:     "a.title",
		DateSelector:      "div.date",
		LocationSelector:  "div.place",
		LinkSelector:      "a.img-wrap",
	}

	afishaConfigTheater := configs.SiteConfig{
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

	// Отправляем сообщение об успешном сбросе конфигураций
	msg := tgbotapi.NewMessage(chatID, "Конфигурации успешно сброшены к дефолтным значениям")
	bot.Send(msg)

	// Пример использования переменных
	fmt.Printf("Конфиг сброшен, ищем: %s", afishaConfigConcert)
	fmt.Printf("Конфиг сброшен, ищем: %s", afishaConfigTheater)
}

func runWebScraper(bot *tgbotapi.BotAPI, chatID int64) {
	// Дефолтная конфигурация для поиска
	afishaConfigConcert := configs.SiteConfig{
		UrlToVisit:        "https://bar.afishagoroda.ru/events/festivali",
		EventType:         "Фестивали",
		AnchestorSelector: "div.events-elem",
		TitleSelector:     "a.title",
		DateSelector:      "div.date",
		LocationSelector:  "div.place",
		LinkSelector:      "a.img-wrap",
	}

	afishaConfigTheater := configs.SiteConfig{
		UrlToVisit:        "https://bar.afishagoroda.ru/events/detiam",
		EventType:         "Детям",
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

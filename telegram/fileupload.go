package telegram

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleFileUpload(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	log.Println("Handling file upload")

	// Проверка, что это ответ на сообщение с запросом конфигурации
	if update.Message.ReplyToMessage != nil {
		log.Printf("Reply to message: %v", update.Message.ReplyToMessage)
	}

	// Получаем файл, отправленный пользователем
	if update.Message.Document == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы не прикрепили файл с конфигурациями :(")
		bot.Send(msg)
		return
	}

	fileID := update.Message.Document.FileID
	log.Printf("File ID: %s", fileID)

	fileInfo, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		log.Printf("Error getting file info: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка получения файла, пожалуйста, попробуйте позже")
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
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка скачивания файла, повторите попытку позднее")
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

	// Разбираем JSON-файл на структуры appconfig.SiteConfig
	err = json.Unmarshal(fileBytes, &userConfigs)
	if err != nil {
		log.Printf("Error unmarshalling JSON: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка разбора файла конфигурации, проверьте ошибки синтаксиса внутри файла")
		bot.Send(msg)
		return
	}

	log.Println("Configurations successfully loaded")

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Конфигурации успешно загружены! Запустите поиск с помощью /run")
	bot.Send(msg)
}

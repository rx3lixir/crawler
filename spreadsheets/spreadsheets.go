package spreadsheets

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rx3lixir/crawler/appconfig"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const (
	googleAuthScope = "https://www.googleapis.com/auth/spreadsheets"
)

// sheetDetails содержит информацию о листе Google Sheets
type sheetDetails struct {
	Id    int64
	Title string
}

// WriteToSpreadsheet записывает события в Google Sheets
func WriteToSpreadsheet(events []appconfig.EventConfig, crawlerAppConfig appconfig.AppConfig) error {
	log.Println("WriteToSpreadsheet called")

	// Декодируем ключ авторизации из base64
	credBytes, err := base64.StdEncoding.DecodeString(crawlerAppConfig.GoogleAuthKey)
	if err != nil {
		return fmt.Errorf("error decoding key JSON: %v", err)
	}

	// Создаем контекст с тайм-аутом
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Создаем конфигурацию для авторизации через JWT
	config, err := google.JWTConfigFromJSON(credBytes, googleAuthScope)
	if err != nil {
		return fmt.Errorf("error creating JWT config: %v", err)
	}

	// Создаем HTTP клиент
	client := config.Client(ctx)

	// Создаем сервис для работы с Google Sheets API
	service, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("error creating Sheets service: %v", err)
	}

	log.Println("Getting data from Google API")
	// Получаем информацию о листах в таблице
	sheetNamesById, err := getSheetNames(service, crawlerAppConfig.SpreadsheetID)
	if err != nil {
		return err
	}

	// Группируем события по типам
	eventGroups := groupEventsByType(events)

	var wg sync.WaitGroup
	errChan := make(chan error, len(eventGroups))

	// Для каждого типа событий запускаем горутину для записи данных в соответствующий лист
	for eventType, events := range eventGroups {
		if details, exists := sheetNamesById[eventType]; exists {
			wg.Add(1)
			go func(eventType string, details sheetDetails, events [][]interface{}) {
				defer wg.Done()
				if err := saveToSheet(service, crawlerAppConfig.SpreadsheetID, details.Title, events); err != nil {
					errChan <- fmt.Errorf("unable to save events to sheet %s: %v", details.Title, err)
				}
				log.Printf("Events saved to sheet %s successfully", details.Title)
			}(eventType, details, events)
		}
	}

	wg.Wait()
	close(errChan)

	// Если были ошибки, собираем их и возвращаем
	if len(errChan) > 0 {
		var errMsg string
		for err := range errChan {
			errMsg += err.Error() + "\n"
		}
		return fmt.Errorf("errors occurred: %v", errMsg)
	}

	return nil
}

// getSheetNames получает имена листов в таблице Google Sheets
func getSheetNames(service *sheets.Service, spreadsheetId string) (map[string]sheetDetails, error) {
	res, err := service.Spreadsheets.Get(spreadsheetId).Fields("sheets(properties(sheetId,title))").Do()
	if err != nil {
		return nil, fmt.Errorf("error getting spreadsheet: %v", err)
	}

	// Создаем мапу, где ключ - название листа, значение - его детали
	sheetNamesById := make(map[string]sheetDetails)
	for _, sheet := range res.Sheets {
		props := sheet.Properties
		sheetNamesById[props.Title] = sheetDetails{Id: props.SheetId, Title: props.Title}
	}

	return sheetNamesById, nil
}

// groupEventsByType группирует события по их типам
func groupEventsByType(events []appconfig.EventConfig) map[string][][]interface{} {
	eventGroups := make(map[string][][]interface{})

	// Для каждого события создаем строку и добавляем в соответствующую группу
	for _, event := range events {
		row := []interface{}{event.Title, event.Date, event.Location, event.Link, event.EventType}
		eventGroups[event.EventType] = append(eventGroups[event.EventType], row)
	}

	return eventGroups
}

// saveToSheet записывает данные в указанный лист Google Sheets
func saveToSheet(service *sheets.Service, spreadsheetId, sheetName string, data [][]interface{}) error {
	log.Printf("saveToSheet called with sheetName: %s", sheetName)
	writeRange := fmt.Sprintf("%s!A1", sheetName)

	valueRange := &sheets.ValueRange{
		Values: data,
	}

	// Обновляем значения в указанном диапазоне листа
	_, err := service.Spreadsheets.Values.Update(spreadsheetId, writeRange, valueRange).ValueInputOption("RAW").Do()
	if err != nil {
		log.Printf("unable to write data to spreadsheet: %v", err)
		return fmt.Errorf("unable to write data to spreadsheet: %v", err)
	}

	log.Printf("Data written to spreadsheet %s in range %s", spreadsheetId, writeRange)
	return nil
}

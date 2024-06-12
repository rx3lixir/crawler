package spreadsheets

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"github.com/rx3lixir/crawler/appconfig"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type sheetDetails struct {
	Id    int64
	Title string
}

func WriteToSpreadsheet(events []appconfig.EventConfig) error {
	log.Println("WriteToSpreadsheet called")

	// Загружаем переменные окружения
	keyJSONBase64 := os.Getenv("GOOGLE_AUTH_KEY")
	spreadSheetId := os.Getenv("SPREADSHEET_ID")

	if keyJSONBase64 == "" || spreadSheetId == "" {
		return fmt.Errorf("GOOGLE_SHEETS_CREDENTIAL_PATH or SPREADSHEET_ID is not set")
	}

	// Декодируем base64 ключ
	credBytes, err := base64.StdEncoding.DecodeString(keyJSONBase64)
	if err != nil {
		return fmt.Errorf("error decoding key JSON: %v", err)
	}

	// Логика аутентификации и инициализации клиента Google Sheets
	ctx := context.Background()
	config, err := google.JWTConfigFromJSON(credBytes, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		return fmt.Errorf("error creating JWT config: %v", err)
	}

	client := config.Client(ctx)

	service, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("error creating Sheets service: %v", err)
	}

	log.Println("Getting data from Google API")

	// Делаем запрос на получение данных с Google Sheets API
	spreadSheetRes, err := service.Spreadsheets.Get(spreadSheetId).Fields("sheets(properties(sheetId,title))").Do()
	if err != nil {
		return fmt.Errorf("error getting spreadsheet: %v", err)
	}

	// Создаем отображение для хранения названий листов по их идентификаторам
	sheetNamesById := make(map[int64]string)

	for _, sheet := range spreadSheetRes.Sheets {
		props := sheet.Properties
		sheetNamesById[props.SheetId] = props.Title
	}

	// Define sheet details for different event types
	sheetMap := map[string]sheetDetails{
		"Концерт":   {Id: 0, Title: sheetNamesById[0]},
		"Театр":     {Id: 434585164, Title: sheetNamesById[434585164]},
		"Фестивали": {Id: 301169124, Title: sheetNamesById[301169124]},
		"Детям":     {Id: 1348865206, Title: sheetNamesById[1348865206]},
	}

	// Group events by type
	eventGroups := map[string][][]interface{}{
		"Концерт":   {},
		"Театр":     {},
		"Фестивали": {},
		"Детям":     {},
	}

	// Фасуем полученные ивенты по группам
	for _, event := range events {
		row := []interface{}{event.Title, event.Date, event.Location, event.Link, event.EventType}
		if _, exists := eventGroups[event.EventType]; exists {
			eventGroups[event.EventType] = append(eventGroups[event.EventType], row)
		}
	}

	// В соответствии названием листа направлям сгруппированные элементы
	for eventType, details := range sheetMap {
		if events, exists := eventGroups[eventType]; exists {
			err := saveToSheet(service, spreadSheetId, details.Title, events)
			if err != nil {
				log.Printf("unable to save events to sheet %s: %v", details.Title, err)
				return fmt.Errorf("unable to save events to sheet %s: %v", details.Title, err)
			}
			log.Printf("Events saved to sheet %s successfully", details.Title)
		}
	}

	return nil
}

func saveToSheet(service *sheets.Service, spreadsheetId, sheetName string, data [][]interface{}) error {
	log.Printf("saveToSheet called with sheetName: %s", sheetName) // Лог в начале функции

	writeRange := fmt.Sprintf("%s!A1", sheetName)

	valueRange := &sheets.ValueRange{
		Values: data,
	}

	_, err := service.Spreadsheets.Values.Update(spreadsheetId, writeRange, valueRange).ValueInputOption("RAW").Do()
	if err != nil {
		log.Printf("unable to write data to spreadsheet: %v", err)
		return fmt.Errorf("unable to write data to spreadsheet: %v", err)
	}

	log.Printf("Data written to spreadsheet %s in range %s", spreadsheetId, writeRange)
	return nil
}

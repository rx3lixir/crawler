package spreadsheets

import (
	"encoding/base64"
	"log"
	"os"

	"github.com/rx3lixir/crawler/config"

	"github.com/joho/godotenv"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func SaveDataToSpreadSheet(events []configs.EventConfig) {
	log.Println("Saving data to spreadsheets:")

	// Booting all env files
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Login logic
	ctx := context.Background()

	credBytes, err := base64.StdEncoding.DecodeString(os.Getenv("KEY_JSON_BASE64"))
	if err != nil {
		log.Fatal(err)
		return
	}

	// Getting config from google API
	config, err := google.JWTConfigFromJSON(credBytes, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatal(err)
		return
	}

	// Initalizing client
	client := config.Client(ctx)

	// Initalizing new spreadsheet service entity
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatal(err)
		return
	}

	spreadSheetId := os.Getenv("GOOGLE_SHEET_ID")

	if spreadSheetId == "" {
		log.Fatal("GOOGLE_SHEET_ID environment variable is not set")
		return
	}

	log.Println("...Getting data from Google API")

	// Делаем запрос на получение данных с Google Sheets API
	spreadSheetRes, err := srv.Spreadsheets.Get(spreadSheetId).Fields("sheets(properties(sheetId,title))").Do()
	if err != nil {
		log.Fatal(err)
		return
	}

	// Создаем отображение для хранения названий листов по их идентификаторам
	sheetNamesById := make(map[int64]string)

	// Проходим по всем листам в ответе от Google Sheets API
	for _, sheet := range spreadSheetRes.Sheets {
		props := sheet.Properties
		sheetNamesById[props.SheetId] = props.Title
	}

	// Идентификаторы листов для "Концерт" и "Театр"
	sheetIdConcert := int64(0)
	sheetIdTheatre := int64(434585164)
	sheetIdFestivali := int64(301169124)
	sheetIdDetiam := int64(1348865206)

	// Получаем названия листов по их идентификаторам
	sheetNameConcert := sheetNamesById[sheetIdConcert]
	sheetNameTheatre := sheetNamesById[sheetIdTheatre]
	sheetNameFestivali := sheetNamesById[sheetIdFestivali]
	sheetNameDetiam := sheetNamesById[sheetIdDetiam]

	// Группируем события по типу
	concertEvents := [][]interface{}{}
	theatreEvents := [][]interface{}{}
	festivalEvents := [][]interface{}{}
	kidsEvents := [][]interface{}{}

	// Фасуем события по срезам
	for _, event := range events {
		row := []interface{}{event.Title, event.Date, event.Location, event.Link, event.EventType}
		switch event.EventType {
		case "Концерт":
			concertEvents = append(concertEvents, row)
		case "Театр":
			theatreEvents = append(theatreEvents, row)
		case "Фестивали":
			festivalEvents = append(festivalEvents, row)
		case "Детям":
			kidsEvents = append(kidsEvents, row)
		}
	}

	// Сохраняем данные на соответствующие листы
	saveToSheet(srv, ctx, spreadSheetId, sheetNameConcert, concertEvents)
	saveToSheet(srv, ctx, spreadSheetId, sheetNameTheatre, theatreEvents)
	saveToSheet(srv, ctx, spreadSheetId, sheetNameDetiam, kidsEvents)
	saveToSheet(srv, ctx, spreadSheetId, sheetNameFestivali, festivalEvents)
}

func saveToSheet(srv *sheets.Service, ctx context.Context, spreadSheetId, sheetName string, values [][]interface{}) {
	log.Printf("Saving data to spreadsheet: %s, sheet: %s", spreadSheetId, sheetName)

	records := sheets.ValueRange{
		Values: values,
	}

	_, err := srv.Spreadsheets.Values.Append(spreadSheetId, sheetName, &records).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Context(ctx).Do()

	if err != nil {
		log.Printf("Error saving data to spreadsheet: %v", err)
		return
	}

	log.Println("Data saved successfully")
}

package spreadsheets

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/rx3lixir/crawler/appconfig"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type sheetDetails struct {
	Id    int64
	Title string
}

func WriteToSpreadsheet(events []appconfig.EventConfig, crawlerAppConfig appconfig.AppConfig) error {
	log.Println("WriteToSpreadsheet called")

	// Получаем данные для авторизации из google API
	credBytes, err := base64.StdEncoding.DecodeString(crawlerAppConfig.GoogleAuthKey)
	if err != nil {
		return fmt.Errorf("error decoding key JSON: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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
	sheetNamesById, err := getSheetNames(service, crawlerAppConfig.SpreadsheetID)
	if err != nil {
		return err
	}

	sheetMap := createSheetMap(sheetNamesById)
	eventGroups := groupEventsByType(events)

	for eventType, details := range sheetMap {
		if events, exists := eventGroups[eventType]; exists {
			err := saveToSheet(service, crawlerAppConfig.SpreadsheetID, details.Title, events)
			if err != nil {
				return fmt.Errorf("unable to save events to sheet %s: %v", details.Title, err)
			}
			log.Printf("Events saved to sheet %s successfully", details.Title)
		}
	}

	return nil
}

func getSheetNames(service *sheets.Service, spreadsheetId string) (map[int64]string, error) {
	res, err := service.Spreadsheets.Get(spreadsheetId).Fields("sheets(properties(sheetId,title))").Do()
	if err != nil {
		return nil, fmt.Errorf("error getting spreadsheet: %v", err)
	}

	sheetNamesById := make(map[int64]string)
	for _, sheet := range res.Sheets {
		props := sheet.Properties
		sheetNamesById[props.SheetId] = props.Title
	}

	return sheetNamesById, nil
}

func createSheetMap(sheetNamesById map[int64]string) map[string]sheetDetails {
	return map[string]sheetDetails{
		"Концерт":   {Id: 0, Title: sheetNamesById[0]},
		"Театр":     {Id: 434585164, Title: sheetNamesById[434585164]},
		"Фестивали": {Id: 301169124, Title: sheetNamesById[301169124]},
		"Детям":     {Id: 1348865206, Title: sheetNamesById[1348865206]},
	}
}

func groupEventsByType(events []appconfig.EventConfig) map[string][][]interface{} {
	eventGroups := map[string][][]interface{}{
		"Концерт":   {},
		"Театр":     {},
		"Фестивали": {},
		"Детям":     {},
	}

	for _, event := range events {
		row := []interface{}{event.Title, event.Date, event.Location, event.Link, event.EventType}
		if _, exists := eventGroups[event.EventType]; exists {
			eventGroups[event.EventType] = append(eventGroups[event.EventType], row)
		}
	}

	return eventGroups
}

func saveToSheet(service *sheets.Service, spreadsheetId, sheetName string, data [][]interface{}) error {
	log.Printf("saveToSheet called with sheetName: %s", sheetName)
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

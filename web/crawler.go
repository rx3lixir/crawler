package web

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/rx3lixir/crawler/appconfig"
)

// WebScraper принимает конфигурации сайтов и возвращает список событий, извлеченных из этих сайтов
func WebScraper(allConfigs []appconfig.SiteConfig) []appconfig.EventConfig {
	var scrapedEvents []appconfig.EventConfig

	var wg sync.WaitGroup // WaitGroup для синхронизации горутин
	var mu sync.Mutex     // Mutex для синхронизации доступа к срезу scrapedEvents

	for _, config := range allConfigs {
		wg.Add(1)
		go func(config appconfig.SiteConfig) {
			defer wg.Done()
			events := extractEvents(config)

			mu.Lock()
			scrapedEvents = append(scrapedEvents, events...)
			mu.Unlock()
		}(config)
	}

	wg.Wait() // Ждем завершения всех горутин

	return scrapedEvents
}

// extractEvents принимает конфигурацию сайта и возвращает список событий, извлеченных из этого сайта
func extractEvents(config appconfig.SiteConfig) []appconfig.EventConfig {
	var extractedEvents []appconfig.EventConfig

	filePath := "rendered_page.html"
	cmd := exec.Command("node", "web/scrape.js", config.UrlToVisit, config.AnchestorSelector, filePath)
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error running Puppeteer script: %v, Output: %s", err, string(cmdOutput))
		return extractedEvents
	}

	htmlContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading rendered HTML: %v", err)
		return extractedEvents
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(htmlContent)))
	if err != nil {
		log.Printf("Error parsing HTML: %v", err)
		return extractedEvents
	}

	doc.Find(config.AnchestorSelector).Each(func(i int, s *goquery.Selection) {
		event, err := extractEventFromElement(config, s)
		if err != nil {
			log.Printf("Error extracting event: %v", err)
			return
		}
		extractedEvents = append(extractedEvents, event)
	})

	// Удаляем файл после обработки
	if err := os.Remove(filePath); err != nil {
		log.Printf("Error removing temp file: %v", err)
	}

	return extractedEvents
}

// extractEventFromElement извлекает событие из HTML элемента в соответствии с конфигурацией сайта
func extractEventFromElement(config appconfig.SiteConfig, element *goquery.Selection) (appconfig.EventConfig, error) {
	// Парсим базовый URL
	baseURL, err := url.Parse(config.UrlToVisit)
	if err != nil {
		return appconfig.EventConfig{}, fmt.Errorf("error parsing base URL: %v", err)
	}

	// Извлекаем href из селектора ссылки
	href, exists := element.Find(config.LinkSelector).Attr("href")

	var fullURL *url.URL
	if exists {
		// Парсим URL из href
		link, err := url.Parse(href)
		if err != nil {
			return appconfig.EventConfig{}, fmt.Errorf("error parsing link URL: %v", err)
		}
		// Создаем полный URL
		fullURL = baseURL.ResolveReference(link)
	}

	// Создаем структуру события
	eventToExtract := appconfig.EventConfig{
		Title:     strings.TrimSpace(element.Find(config.TitleSelector).Text()),
		Date:      strings.TrimSpace(element.Find(config.DateSelector).Text()),
		Location:  config.LocationSelector, // strings.TrimSpace(element.Find(config.LocationSelector).Text()),
		Link:      config.UrlToVisit,       // Устанавливаем пустой URL по умолчанию
		EventType: config.EventType,
	}

	if fullURL != nil {
		eventToExtract.Link = fullURL.String()
	}

	return eventToExtract, nil
}

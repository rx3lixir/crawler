package web

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"

	"github.com/gocolly/colly"
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

	// Создаем colly коллектора
	c := colly.NewCollector()

	var wg sync.WaitGroup // WaitGroup для синхронизации горутин

	c.OnHTML(config.AnchestorSelector, func(element *colly.HTMLElement) {
		wg.Add(1)
		go func() {
			defer wg.Done()

			event, err := extractEventFromElement(config, element)
			if err != nil {
				log.Printf("Error extracting event: %v", err)
				return
			}

			// Mutex для синхронизации доступа к срезу extractedEvents
			mu := sync.Mutex{}
			mu.Lock()
			extractedEvents = append(extractedEvents, event)
			mu.Unlock()
		}()
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting: ", r.URL)
	})

	err := c.Visit(config.UrlToVisit)
	if err != nil {
		log.Printf("Error visiting URL: %v", err)
	}

	wg.Wait() // Ждем завершения всех горутин

	return extractedEvents
}

// extractEventFromElement извлекает событие из HTML элемента в соответствии с конфигурацией сайта
func extractEventFromElement(config appconfig.SiteConfig, element *colly.HTMLElement) (appconfig.EventConfig, error) {
	// Получаем DOM элемент
	elemDOM := element.DOM

	// Парсим базовый URL
	baseURL, err := url.Parse(config.UrlToVisit)
	if err != nil {
		return appconfig.EventConfig{}, fmt.Errorf("error parsing base URL: %v", err)
	}

	// Извлекаем href из селектора ссылки
	href, exists := elemDOM.Find(config.LinkSelector).Attr("href")
	if !exists {
		return appconfig.EventConfig{}, fmt.Errorf("no href found for %s", elemDOM.Find(config.TitleSelector).Text())
	}

	// Парсим URL из href
	link, err := url.Parse(href)
	if err != nil {
		return appconfig.EventConfig{}, fmt.Errorf("error parsing link URL: %v", err)
	}

	// Создаем полный URL
	fullURL := baseURL.ResolveReference(link)

	// Создаем структуру события
	eventToExtract := appconfig.EventConfig{
		Title:     elemDOM.Find(config.TitleSelector).Text(),
		Date:      elemDOM.Find(config.DateSelector).Text(),
		Location:  strings.TrimSpace(elemDOM.Find(config.LocationSelector).Text()),
		Link:      fullURL.String(),
		EventType: config.EventType,
	}

	return eventToExtract, nil
}

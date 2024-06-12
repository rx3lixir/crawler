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

func WebScraper(allConfigs []appconfig.SiteConfig) []appconfig.EventConfig {
	scrapedEvents := []appconfig.EventConfig{}

	for _, config := range allConfigs {
		events := extractEvents(config)

		scrapedEvents = append(scrapedEvents, events...)
	}

	return scrapedEvents
}

func extractEvents(config appconfig.SiteConfig) []appconfig.EventConfig {
	// A slice for all storing events
	var extractedEvents []appconfig.EventConfig

	// Creating colly entity
	c := colly.NewCollector()

	var wg sync.WaitGroup // Создаем WaitGroup для синхронизации горутин

	c.OnHTML(config.AnchestorSelector, func(element *colly.HTMLElement) {
		wg.Add(1) // Увеличиваем счетчик WaitGroup перед созданием горутины

		go func() {
			defer wg.Done() // Уменьшаем счетчик WaitGroup после выполнения горутины

			event, err := extractEventFromElement(config, element)
			if err != nil {
				log.Printf("Error extracting event: %v", err)
				return
			}

			extractedEvents = append(extractedEvents, event)
		}()
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting: ", r.URL)
	})

	err := c.Visit(config.UrlToVisit)
	if err != nil {
		log.Fatal(err)
	}

	wg.Wait() // Ждем завершения всех горутин

	return extractedEvents
}

func extractEventFromElement(config appconfig.SiteConfig, element *colly.HTMLElement) (appconfig.EventConfig, error) {
	// Initializing entity for getting DOM elements
	elemDOM := element.DOM

	// Getting base URL
	baseURL, err := url.Parse(config.UrlToVisit)
	if err != nil {
		return appconfig.EventConfig{}, fmt.Errorf("error parsing base URL: %v", err)
	}

	// Getting href from link selector
	href, exists := elemDOM.Find(config.LinkSelector).Attr("href")
	if !exists {
		return appconfig.EventConfig{}, fmt.Errorf("no href found for %s", elemDOM.Find(config.TitleSelector).Text())
	}

	// Parsing url from href attr
	link, err := url.Parse(href)
	if err != nil {
		return appconfig.EventConfig{}, fmt.Errorf("error parsing link URL: %v", err)
	}

	// Resolving reference for full URl
	fullURL := baseURL.ResolveReference(link)

	eventToExtract := appconfig.EventConfig{
		Title:     elemDOM.Find(config.TitleSelector).Text(),
		Date:      elemDOM.Find(config.DateSelector).Text(),
		Location:  strings.TrimSpace(elemDOM.Find(config.LocationSelector).Text()),
		Link:      fullURL.String(),
		EventType: config.EventType,
	}

	return eventToExtract, nil
}

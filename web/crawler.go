package web

import (
	"log"
	"net/url"
	"strings"

	"github.com/gocolly/colly"
	"github.com/rx3lixir/crawler/config"
)

func WebScraper(allConfigs []configs.SiteConfig) []configs.EventConfig {
	scrapedEvents := []configs.EventConfig{}

	for _, config := range allConfigs {
		events := extractEvents(config)

		scrapedEvents = append(scrapedEvents, events...)
	}

	return scrapedEvents
}

func extractEvents(config configs.SiteConfig) []configs.EventConfig {
	var extractedEvents []configs.EventConfig

	c := colly.NewCollector()

	c.OnHTML(config.AnchestorSelector, func(element *colly.HTMLElement) {
		baseURL, err := url.Parse(config.UrlToVisit)

		if err != nil {
			log.Printf("Error parsing base URL: %v", err)
		}

		elemDOM := element.DOM

		href, exists := elemDOM.Find(config.LinkSelector).Attr("href")

		if exists {
			link, err := url.Parse(href)
			if err != nil {
				log.Printf("Error parsing link URL: %v", err)
			} else {
				fullURL := baseURL.ResolveReference(link)
				eventToExtract := configs.EventConfig{
					Title:     elemDOM.Find(config.TitleSelector).Text(),
					Date:      elemDOM.Find(config.DateSelector).Text(),
					Location:  strings.TrimSpace(elemDOM.Find(config.LocationSelector).Text()),
					Link:      fullURL.String(),
					EventType: config.EventType,
				}
				extractedEvents = append(extractedEvents, eventToExtract)
			}
		} else {
			log.Printf("No href found for %s", elemDOM.Find(config.TitleSelector).Text())
		}
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting: ", r.URL)
	})

	err := c.Visit(config.UrlToVisit)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(extractedEvents)
	return extractedEvents
}

package main

import (
	"log"
	"net/url"
	"strings"

	"github.com/gocolly/colly"
)

type Event struct {
	Title     string `json:"title"`
	Date      string `json:"date"`
	Location  string `json:"location"`
	Link      string `json:"link"`
	EventType string `json:"eventType"`
}

type SiteConfig struct {
	UrlToVisit        string
	EventType         string
	AnchestorSelector string
	TitleSelector     string
	DateSelector      string
	LocationSelector  string
	LinkSelector      string
}

func extractEvents(config SiteConfig) []Event {
	var extractedEvents []Event

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
				eventToExtract := Event{
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

func WebScraper(configs []SiteConfig) []Event {
	scrapedEvents := []Event{}

	for _, config := range configs {
		events := extractEvents(config)

		scrapedEvents = append(scrapedEvents, events...)
	}

	return scrapedEvents
}

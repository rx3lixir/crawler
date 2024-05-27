package main

import (
	"log"
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
	UrlToVisit       string
	EventType        string
	TitleSelector    string
	DateSelector     string
	LocationSelector string
	LinkSelector     string
}

func extractEvents(config SiteConfig) []Event {
	var extractedEvents []Event

	c := colly.NewCollector()

	c.OnHTML(config.TitleSelector, func(element *colly.HTMLElement) {
		elemDOM := element.DOM

		eventToExtract := Event{
			Title:     elemDOM.Find(config.TitleSelector).Text(),
			Date:      elemDOM.Find(config.DateSelector).Text(),
			Location:  strings.TrimSpace(elemDOM.Find(config.LocationSelector).Text()),
			Link:      config.UrlToVisit,
			EventType: config.EventType,
		}

		log.Println(eventToExtract)

		extractedEvents = append(extractedEvents, eventToExtract)
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting: ", r.URL)
	})

	err := c.Visit(config.UrlToVisit)
	if err != nil {
		log.Fatal(err)
	}

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

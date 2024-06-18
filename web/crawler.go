package web

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/url"
	"os/exec"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/rx3lixir/crawler/appconfig"
	"golang.org/x/sync/semaphore"
)

const maxConcurrentScrapes = 5

// WebScraper принимает конфигурации сайтов и возвращает список событий, извлеченных из этих сайтов
func WebScraper(allConfigs []appconfig.SiteConfig) []appconfig.EventConfig {
	var scrapedEvents []appconfig.EventConfig
	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := semaphore.NewWeighted(maxConcurrentScrapes)

	for _, config := range allConfigs {
		wg.Add(1)
		go func(config appconfig.SiteConfig) {
			defer wg.Done()
			if err := sem.Acquire(context.Background(), 1); err != nil {
				log.Printf("Failed to acquire semaphore: %v", err)
				return
			}
			defer sem.Release(1)
			log.Printf("Starting extraction for site: %s", config.UrlToVisit)
			events := extractEvents(config)

			mu.Lock()
			scrapedEvents = append(scrapedEvents, events...)
			mu.Unlock()
			log.Printf("Finished extraction for site: %s", config.UrlToVisit)
		}(config)
	}

	wg.Wait()

	return scrapedEvents
}

// extractEvents принимает конфигурацию сайта и возвращает список событий, извлеченных из этого сайта
func extractEvents(config appconfig.SiteConfig) []appconfig.EventConfig {
	var extractedEvents []appconfig.EventConfig

	cmd := exec.Command("node", "web/scrape.js", config.UrlToVisit, config.AnchestorSelector)
	var cmdOutput bytes.Buffer
	cmd.Stdout = &cmdOutput
	cmd.Stderr = &cmdOutput

	log.Printf("Running Puppeteer script for URL: %s", config.UrlToVisit)
	err := cmd.Run()
	if err != nil {
		log.Printf("Error running Puppeteer script: %v, Output: %s", err, cmdOutput.String())
		return extractedEvents
	}

	htmlContent := cmdOutput.String()
	if htmlContent == "" {
		log.Printf("No HTML content fetched for URL: %s", config.UrlToVisit)
		return extractedEvents
	}

	log.Printf("HTML content fetched for URL: %s", config.UrlToVisit)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
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

	return extractedEvents
}

// extractEventFromElement извлекает событие из HTML элемента в соответствии с конфигурацией сайта
func extractEventFromElement(config appconfig.SiteConfig, element *goquery.Selection) (appconfig.EventConfig, error) {
	baseURL, err := url.Parse(config.UrlToVisit)
	if err != nil {
		return appconfig.EventConfig{}, fmt.Errorf("error parsing base URL: %v", err)
	}

	href, exists := element.Find(config.LinkSelector).Attr("href")

	var fullURL *url.URL
	if exists {
		link, err := url.Parse(href)
		if err != nil {
			return appconfig.EventConfig{}, fmt.Errorf("error parsing link URL: %v", err)
		}
		fullURL = baseURL.ResolveReference(link)
	}

	eventToExtract := appconfig.EventConfig{
		Title:     strings.TrimSpace(element.Find(config.TitleSelector).Text()),
		Date:      strings.TrimSpace(element.Find(config.DateSelector).Text()),
		Location:  config.LocationSelector,
		Link:      config.UrlToVisit,
		EventType: config.EventType,
	}

	if fullURL != nil {
		eventToExtract.Link = fullURL.String()
	}

	return eventToExtract, nil
}

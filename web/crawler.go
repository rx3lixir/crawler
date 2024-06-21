package web

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/rx3lixir/crawler/appconfig"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
)

const maxConcurrentScrapes = 15

var log = logrus.New()

func init() {
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})
	log.SetLevel(logrus.InfoLevel)
}

// WebScraper processes site configurations and returns a list of extracted events.
func WebScraper(allConfigs []appconfig.SiteConfig) []appconfig.EventConfig {
	var scrapedEvents []appconfig.EventConfig
	var wg sync.WaitGroup
	var mu sync.Mutex

	sem := semaphore.NewWeighted(maxConcurrentScrapes)
	ctx := context.Background()

	for _, config := range allConfigs {
		wg.Add(1)
		go func(config appconfig.SiteConfig) {
			defer wg.Done()
			if err := sem.Acquire(ctx, 1); err != nil {
				log.Errorf("Failed to acquire semaphore: %v", err)
				return
			}
			defer sem.Release(1)

			log.Infof("Starting extraction for site: %s", config.UrlToVisit)
			events := extractEvents(config)

			mu.Lock()
			scrapedEvents = append(scrapedEvents, events...)
			mu.Unlock()
			log.Infof("Finished extraction for site: %s", config.UrlToVisit)
		}(config)
	}

	wg.Wait()

	return scrapedEvents
}

// extractEvents extracts events from a given site configuration.
func extractEvents(config appconfig.SiteConfig) []appconfig.EventConfig {
	var extractedEvents []appconfig.EventConfig

	for retries := 0; retries < 3; retries++ {
		cmd := exec.Command("node", "web/scrape.js", config.UrlToVisit, config.AnchestorSelector)
		var cmdOutput bytes.Buffer
		cmd.Stdout = &cmdOutput
		cmd.Stderr = &cmdOutput

		log.Infof("Running Puppeteer script for URL: %s", config.UrlToVisit)
		err := cmd.Run()
		if err != nil {
			log.Errorf("Error running Puppeteer script: %v, Output: %s", err, cmdOutput.String())
			if retries < 2 {
				log.Infof("Retrying... (%d/3)", retries+1)
				time.Sleep(2 * time.Second)
				continue
			}
			return extractedEvents
		}

		htmlContent := cmdOutput.String()
		if htmlContent == "" {
			log.Warnf("No HTML content fetched for URL: %s", config.UrlToVisit)
			return extractedEvents
		}

		log.Infof("HTML content fetched for URL: %s", config.UrlToVisit)

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
		if err != nil {
			log.Errorf("Error parsing HTML: %v", err)
			return extractedEvents
		}

		doc.Find(config.AnchestorSelector).Each(func(i int, s *goquery.Selection) {
			log.Infof("Found element with selector %s on URL: %s", config.AnchestorSelector, config.UrlToVisit)
			event, err := extractEventFromElement(config, s)
			if err != nil {
				log.Errorf("Error extracting event: %v", err)
				return
			}
			log.Infof("Extracted event: %+v", event)
			extractedEvents = append(extractedEvents, event)
		})

		return extractedEvents
	}

	return extractedEvents
}

// extractEventFromElement extracts an event from a HTML element based on site configuration.
func extractEventFromElement(config appconfig.SiteConfig, element *goquery.Selection) (appconfig.EventConfig, error) {
	baseURL, err := url.Parse(config.UrlToVisit)
	if err != nil {
		return appconfig.EventConfig{}, fmt.Errorf("error parsing base URL: %v", err)
	}

	href, exists := element.Find(config.LinkSelector).Attr("href")
	log.Infof("Link found: %v (exists: %v)", href, exists)

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

	log.Infof("Extracted event details: %+v", eventToExtract)

	return eventToExtract, nil
}

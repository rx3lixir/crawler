package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/rx3lixir/crawler/appconfig"
	"github.com/sirupsen/logrus"
)

const (
	maxWorkers = 20
	maxRetries = 3
)

var log = logrus.New()

func init() {
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})
	log.SetLevel(logrus.InfoLevel)
}

type Job struct {
	config appconfig.SiteConfig
}

type Result struct {
	events []appconfig.EventConfig
	err    error
}

// WebScraper processes site configurations and returns a list of extracted events.
func WebScraper(allConfigs []appconfig.SiteConfig) []appconfig.EventConfig {
	jobs := make(chan Job, len(allConfigs))
	results := make(chan Result, len(allConfigs))

	// Start worker pool
	for i := 0; i < maxWorkers; i++ {
		go worker(jobs, results)
	}

	// Send jobs to the pool
	for _, config := range allConfigs {
		jobs <- Job{config: config}
	}
	close(jobs)

	// Collect results
	var scrapedEvents []appconfig.EventConfig
	for i := 0; i < len(allConfigs); i++ {
		result := <-results
		if result.err != nil {
			log.Errorf("Error processing job: %v", result.err)
		} else {
			scrapedEvents = append(scrapedEvents, result.events...)
		}
	}

	return scrapedEvents
}

func worker(jobs <-chan Job, results chan<- Result) {
	for job := range jobs {
		log.Infof("Starting extraction for site: %s", job.config.UrlToVisit)
		events, err := extractEvents(job.config)
		if err != nil {
			results <- Result{err: err}
		} else {
			results <- Result{events: events}
		}
		log.Infof("Finished extraction for site: %s", job.config.UrlToVisit)
	}
}

func extractEvents(config appconfig.SiteConfig) ([]appconfig.EventConfig, error) {
	var extractedEvents []appconfig.EventConfig

	for retries := 0; retries < maxRetries; retries++ {
		// Prepare the request to the Puppeteer service
		reqBody, err := json.Marshal(map[string]string{
			"url":      config.UrlToVisit,
			"selector": config.AnchestorSelector,
		})
		if err != nil {
			return nil, fmt.Errorf("error preparing request: %v", err)
		}

		resp, err := http.Post("http://localhost:3000/scrape", "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			log.Errorf("Error making request to Puppeteer service: %v", err)
			if retries < maxRetries-1 {
				log.Infof("Retrying... (%d/%d)", retries+1, maxRetries)
				time.Sleep(2 * time.Second)
				continue
			}
			return nil, fmt.Errorf("failed to scrape after %d retries", maxRetries)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response: %v", err)
		}

		var result struct {
			HTML string `json:"html"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("error parsing response: %v", err)
		}

		// Process the HTML content as before
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(result.HTML))
		if err != nil {
			return nil, fmt.Errorf("error parsing HTML: %v", err)
		}

		doc.Find(config.AnchestorSelector).Each(func(i int, s *goquery.Selection) {
			event, err := extractEventFromElement(config, s)
			if err != nil {
				log.Errorf("Error extracting event: %v", err)
				return
			}
			extractedEvents = append(extractedEvents, event)
		})

		return extractedEvents, nil
	}

	return nil, fmt.Errorf("failed to extract events after %d retries", maxRetries)
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

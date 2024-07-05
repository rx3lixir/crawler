package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/rx3lixir/crawler/appconfig"
	"github.com/sirupsen/logrus"
)

var (
	log        = logrus.New()
	httpClient = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 20,
		},
	}
)

func init() {
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})
	log.SetLevel(logrus.InfoLevel)
}

// WebScraper processes site configurations and returns a list of extracted events.
func WebScraper(crawlerAppConfig appconfig.AppConfig, ctx context.Context, allConfigs []appconfig.SiteConfig) []appconfig.EventConfig {

	jobs := make(chan appconfig.Job, len(allConfigs))
	results := make(chan appconfig.Result, len(allConfigs))

	workerCount := crawlerAppConfig.MaxWorkers

	// Start worker pool
	for i := 0; i < workerCount; i++ {
		go worker(crawlerAppConfig, ctx, jobs, results)
	}

	// Send jobs to the pool
	for _, config := range allConfigs {
		jobs <- appconfig.Job{Config: config}
	}
	close(jobs)

	// Collect results
	var scrapedEvents []appconfig.EventConfig
	for i := 0; i < len(allConfigs); i++ {
		select {
		case result := <-results:
			if result.Err != nil {
				log.Errorf("Error processing job: %v", result.Err)
			} else {
				scrapedEvents = append(scrapedEvents, result.Events...)
			}
		case <-ctx.Done():
			log.Warn("Context cancelled, stopping web scraper")
			return scrapedEvents
		}
	}

	return scrapedEvents
}

func worker(crawlerAppConfig appconfig.AppConfig, ctx context.Context, jobs <-chan appconfig.Job, results chan<- appconfig.Result) {
	for job := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
			log.Infof("Starting extraction for site: %s", job.Config.UrlToVisit)
			events, err := extractEvents(crawlerAppConfig, ctx, job.Config)
			results <- appconfig.Result{Events: events, Err: err}
			log.Infof("Finished extraction for site: %s", job.Config.UrlToVisit)
		}
	}
}

func extractEvents(crawlerAppConfig appconfig.AppConfig, ctx context.Context, config appconfig.SiteConfig) ([]appconfig.EventConfig, error) {
	var extractedEvents []appconfig.EventConfig

	html, err := fetchHTMLWithRetry(crawlerAppConfig, ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch HTML: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %w", err)
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

func fetchHTMLWithRetry(crawlerAppConfig appconfig.AppConfig, ctx context.Context, config appconfig.SiteConfig) (string, error) {
	var html string
	var err error

	for retries := 0; retries < int(crawlerAppConfig.MaxRetries); retries++ {
		html, err = fetchHTML(ctx, config)
		if err == nil {
			return html, nil
		}

		log.Warnf("Error fetching HTML (attempt %d/%d): %v", retries+1, appconfig.CrawlerApp.MaxRetries, err)
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(time.Duration(retries+1) * time.Second):
			// Exponential backoff
		}
	}

	return "", fmt.Errorf("failed to fetch HTML after %d retries: %w", appconfig.CrawlerApp.MaxRetries, err)
}

func fetchHTML(ctx context.Context, config appconfig.SiteConfig) (string, error) {
	reqBody, err := json.Marshal(map[string]string{
		"url":      config.UrlToVisit,
		"selector": config.AnchestorSelector,
	})
	if err != nil {
		return "", fmt.Errorf("error preparing request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:3000/scrape", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request to Puppeteer service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	var result struct {
		HTML string `json:"html"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error parsing response: %w", err)
	}

	return result.HTML, nil
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

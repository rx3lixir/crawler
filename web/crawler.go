package web

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/rx3lixir/crawler/appconfig"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

var (
	log = logrus.New()
)

func init() {
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})
	log.SetLevel(logrus.InfoLevel)
}

func WebScraper(crawlerAppConfig appconfig.AppConfig, ctx context.Context, allConfigs []appconfig.SiteConfig) []appconfig.EventConfig {
	log.Printf("WebScraper started with %d configs", len(allConfigs))

	jobs := make(chan appconfig.Job, len(allConfigs))
	results := make(chan appconfig.Result, len(allConfigs))

	workerCount := crawlerAppConfig.MaxWorkers
	if workerCount == 0 {
		workerCount = len(allConfigs)
	}
	log.Printf("Using %d workers", workerCount)

	limiter := rate.NewLimiter(rate.Every(time.Second), 5)

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(crawlerAppConfig, ctx, jobs, results, limiter, &wg)
	}

	log.Printf("Sending jobs to workers")
	for _, config := range allConfigs {
		jobs <- appconfig.Job{Config: config}
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
		log.Printf("All workers finished")
	}()

	var scrapedEvents []appconfig.EventConfig

	log.Printf("Collecting results")
	for result := range results {
		if result.Err != nil {
			log.Warnf("Error processing job: %v", result.Err)
		} else {
			scrapedEvents = append(scrapedEvents, result.Events...)
		}
	}

	log.Printf("WebScraper finished, scraped %d events", len(scrapedEvents))
	return scrapedEvents
}

func worker(crawlerAppConfig appconfig.AppConfig, ctx context.Context, jobs <-chan appconfig.Job, results chan<- appconfig.Result, limiter *rate.Limiter, wg *sync.WaitGroup) {
	defer wg.Done()

	// Create a new browser context for each worker
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	browserCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	for job := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
			if err := limiter.Wait(ctx); err != nil {
				results <- appconfig.Result{Err: err}
				continue
			}
			log.Debugf("Starting extraction for site: %s", job.Config.UrlToVisit)
			events, err := extractEvents(crawlerAppConfig, browserCtx, job.Config)
			results <- appconfig.Result{Events: events, Err: err}
			log.Debugf("Finished extraction for site: %s", job.Config.UrlToVisit)
		}
	}
}

func extractEvents(crawlerAppConfig appconfig.AppConfig, ctx context.Context, config appconfig.SiteConfig) ([]appconfig.EventConfig, error) {
	var extractedEvents []appconfig.EventConfig

	html, err := fetchHTMLWithChromedp(ctx, config.UrlToVisit, config.AnchestorSelector)
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
			log.Warnf("Error extracting event: %v", err)
			return
		}
		extractedEvents = append(extractedEvents, event)
	})

	return extractedEvents, nil
}

func fetchHTMLWithChromedp(ctx context.Context, urlToVisit, selector string) (string, error) {
	var htmlContent string

	err := chromedp.Run(ctx,
		chromedp.Navigate(urlToVisit),
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		return "", fmt.Errorf("failed to navigate and extract HTML: %w", err)
	}

	return htmlContent, nil
}

func extractEventFromElement(config appconfig.SiteConfig, element *goquery.Selection) (appconfig.EventConfig, error) {
	baseURL, err := url.Parse(config.UrlToVisit)
	if err != nil {
		return appconfig.EventConfig{}, fmt.Errorf("error parsing base URL: %v", err)
	}

	href, exists := element.Find(config.LinkSelector).Attr("href")
	log.Debugf("Link found: %v (exists: %v)", href, exists)

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

	log.Debugf("Extracted event details: %+v", eventToExtract)

	return eventToExtract, nil
}

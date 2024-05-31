package main

import (
	"github.com/rx3lixir/crawler/config"
	"github.com/rx3lixir/crawler/spreadsheets"
	"github.com/rx3lixir/crawler/web"
)

func main() {

	afishaConfigConcert := configs.SiteConfig{
		UrlToVisit:        "https://bar.afishagoroda.ru/events/koncert",
		EventType:         "Концерт",
		AnchestorSelector: "div.events-elem",
		TitleSelector:     "a.title",
		DateSelector:      "div.date",
		LocationSelector:  "div.place",
		LinkSelector:      "a.img-wrap",
	}

	afishaConfigTheater := configs.SiteConfig{
		UrlToVisit:        "https://bar.afishagoroda.ru/events/teatr",
		EventType:         "Театр",
		AnchestorSelector: "div.events-elem",
		TitleSelector:     "a.title",
		DateSelector:      "div.date",
		LocationSelector:  "div.place",
		LinkSelector:      "a.img-wrap",
	}

	SearchSitesConfigs := []configs.SiteConfig{afishaConfigConcert, afishaConfigTheater}

	allEvents := web.WebScraper(SearchSitesConfigs)

	spreadsheets.SaveDataToSpreadSheet(allEvents)

}

package main

func main() {

	afishaConfigConcert := SiteConfig{
		UrlToVisit:        "https://bar.afishagoroda.ru/events/koncert",
		EventType:         "Концерт",
		AnchestorSelector: "div.events-elem",
		TitleSelector:     "a.title",
		DateSelector:      "div.date",
		LocationSelector:  "div.place",
		LinkSelector:      "a.img-wrap",
	}

	afishaConfigTheater := SiteConfig{
		UrlToVisit:        "https://bar.afishagoroda.ru/events/teatr",
		EventType:         "Театр",
		AnchestorSelector: "div.events-elem",
		TitleSelector:     "a.title",
		DateSelector:      "div.date",
		LocationSelector:  "div.place",
		LinkSelector:      "a.img-wrap",
	}

	SearchSitesConfigs := []SiteConfig{afishaConfigConcert, afishaConfigTheater}

	allEvents := WebScraper(SearchSitesConfigs)

	saveDataToSpreadSheet(allEvents)
}

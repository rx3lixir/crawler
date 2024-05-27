package main

func main() {

	afishaConfigConcert := SiteConfig{
		UrlToVisit:       "https://bar.afishagoroda.ru/events/koncert",
		EventType:        "Концерт",
		TitleSelector:    "a.title",
		DateSelector:     "div.date",
		LocationSelector: "div.place a",
		LinkSelector:     "a.title",
	}

	afishaConfigTheater := SiteConfig{
		UrlToVisit:       "https://bar.afishagoroda.ru/events/teatr",
		EventType:        "Театр",
		TitleSelector:    "a.title",
		DateSelector:     "div.date",
		LocationSelector: "div.place a",
		LinkSelector:     "a.title",
	}

	SearchSitesConfigs := []SiteConfig{afishaConfigConcert, afishaConfigTheater}

	allEvents := WebScraper(SearchSitesConfigs)

	saveDataToSpreadSheet(allEvents)
}

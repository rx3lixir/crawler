package configs

type SiteConfig struct {
	UrlToVisit        string
	EventType         string
	AnchestorSelector string
	TitleSelector     string
	DateSelector      string
	LocationSelector  string
	LinkSelector      string
}

type EventConfig struct {
	Title     string `json:"title"`
	Date      string `json:"date"`
	Location  string `json:"location"`
	Link      string `json:"link"`
	EventType string `json:"eventType"`
}

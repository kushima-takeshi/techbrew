package model

import "time"

type Article struct {
	Title     string
	URL       string
	Summary   string
	Published time.Time
	Source    string
}

type Topic struct {
	Title       string   `json:"title"`
	Summary     string   `json:"summary"`
	RelatedURLs []string `json:"related_urls"`
}

type Digest struct {
	Topics   []Topic
	Articles []Article
}

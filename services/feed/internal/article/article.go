package article

import "time"

type Article struct {
	ID          string    `json:"id"`
	Source      string    `json:"source"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	Tags        []string  `json:"tags"`
	PublishedAt time.Time `json:"published_at"`
}

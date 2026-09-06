package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type Article struct {
	ID                 string    `json:"id"`
	Source             string    `json:"source"`
	URL                string    `json:"url"`
	Title              string    `json:"title"`
	Summary            string    `json:"summary"`
	TitleHighlighted   string    `json:"title_highlighted,omitempty"`
	SummaryHighlighted string    `json:"summary_highlighted,omitempty"`
	Tags               []string  `json:"tags"`
	PublishedAt        time.Time `json:"published_at"`
}

func IDFromURL(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:16])
}

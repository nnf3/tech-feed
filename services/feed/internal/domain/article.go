package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

const (
	ArticleKindTutorial = "tutorial"
	ArticleKindNews     = "news"
	ArticleKindOpinion  = "opinion"
	ArticleKindRelease  = "release"
	ArticleKindOther    = "other"

	ArticleLevelBeginner     = "beginner"
	ArticleLevelIntermediate = "intermediate"
	ArticleLevelAdvanced     = "advanced"
)

type Article struct {
	ID                 string     `json:"id"`
	Source             string     `json:"source"`
	URL                string     `json:"url"`
	Title              string     `json:"title"`
	Summary            string     `json:"summary"`
	TitleHighlighted   string     `json:"title_highlighted,omitempty"`
	SummaryHighlighted string     `json:"summary_highlighted,omitempty"`
	Tags               []string   `json:"tags"`
	Kind               string     `json:"kind,omitempty"`
	Level              string     `json:"level,omitempty"`
	Quality            *float64   `json:"quality,omitempty"`
	Promo              bool       `json:"promo,omitempty"`
	EnrichedAt         *time.Time `json:"enriched_at,omitempty"`
	PublishedAt        time.Time  `json:"published_at"`
}

// CopySignals は TypeSafe 由来の判定だけをコピーする。本文やタグは触らない。
func (a *Article) CopySignals(from Article) {
	a.Kind = from.Kind
	a.Level = from.Level
	a.Quality = from.Quality
	a.Promo = from.Promo
	a.EnrichedAt = from.EnrichedAt
}

// NeedsEnrichment は TypeSafe にまだ判定させていない記事かを返す。
// enriched_at が無い旧データも、判定が1つでも付いていれば判定済みとみなす。
func (a Article) NeedsEnrichment() bool {
	if a.EnrichedAt != nil {
		return false
	}
	return a.Kind == "" && a.Level == "" && a.Quality == nil && !a.Promo
}

func IDFromURL(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:16])
}

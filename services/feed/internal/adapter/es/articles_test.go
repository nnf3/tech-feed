package es

import (
	"testing"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

func TestFeedPageFromHits(t *testing.T) {
	hits := make([]articleHit, domain.FeedPageSize+1)
	for i := range hits {
		hits[i] = articleHit{
			Source: domain.Article{ID: string(rune('a' + i))},
			Sort:   []any{float64(i), string(rune('a' + i))},
		}
	}
	page, err := feedPageFromHits(hits)
	if err != nil || len(page.Articles) != domain.FeedPageSize || page.Next == "" {
		t.Fatalf("%#v %v", page, err)
	}
	short, err := feedPageFromHits(hits[:3])
	if err != nil || len(short.Articles) != 3 || short.Next != "" {
		t.Fatalf("short: %#v %v", short, err)
	}
}

func TestApplyHighlight(t *testing.T) {
	got := applyHighlight(domain.Article{Title: "Go", Summary: "lang"}, map[string][]string{
		"title":   {"<mark>Go</mark>"},
		"summary": {"<mark>lang</mark>"},
	})
	if got.TitleHighlighted != "<mark>Go</mark>" || got.SummaryHighlighted != "<mark>lang</mark>" {
		t.Fatalf("%#v", got)
	}
	plain := applyHighlight(domain.Article{Title: "Go"}, nil)
	if plain.TitleHighlighted != "" {
		t.Fatalf("empty highlight leaked: %#v", plain)
	}
}

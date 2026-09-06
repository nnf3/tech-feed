package es

import (
	"testing"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

func TestArticleSearchBodyFilterTags(t *testing.T) {
	body := articleSearchBody(domain.FeedQuery{FilterTags: []string{"react"}})
	query, _ := body["query"].(map[string]any)
	boolQuery, _ := query["bool"].(map[string]any)
	filters, ok := boolQuery["filter"].([]any)
	if !ok || len(filters) == 0 {
		t.Fatalf("filter missing: %#v", boolQuery)
	}
	terms, _ := filters[0].(map[string]any)["terms"].(map[string]any)
	tags, _ := terms["tags"].([]string)
	if len(tags) != 1 || tags[0] != "react" {
		t.Fatalf("tags: %#v", terms["tags"])
	}
}

func TestArticleSearchBodyHighlight(t *testing.T) {
	if _, ok := articleSearchBody(domain.FeedQuery{})["highlight"]; ok {
		t.Fatal("highlight should be off without text")
	}
	body := articleSearchBody(domain.FeedQuery{Text: "go"})
	highlight, ok := body["highlight"].(map[string]any)
	if !ok {
		t.Fatal("highlight missing")
	}
	if highlight["encoder"] != "html" {
		t.Fatalf("encoder: %#v", highlight["encoder"])
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

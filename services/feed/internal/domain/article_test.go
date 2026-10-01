package domain

import (
	"testing"
	"time"
)

func TestNeedsEnrichment(t *testing.T) {
	q := 0.5
	at := time.Now()
	cases := []struct {
		name    string
		article Article
		want    bool
	}{
		{"fresh", Article{ID: "a"}, true},
		{"enriched with no confident answers", Article{ID: "a", EnrichedAt: &at}, false},
		{"legacy kind", Article{ID: "a", Kind: ArticleKindNews}, false},
		{"legacy level", Article{ID: "a", Level: ArticleLevelBeginner}, false},
		{"legacy quality", Article{ID: "a", Quality: &q}, false},
		{"legacy promo", Article{ID: "a", Promo: true}, false},
	}
	for _, tc := range cases {
		if got := tc.article.NeedsEnrichment(); got != tc.want {
			t.Errorf("%s: got %v want %v", tc.name, got, tc.want)
		}
	}
}

func TestCopySignalsLeavesContent(t *testing.T) {
	dst := Article{ID: "a", Title: "new", Tags: []string{"go"}}
	src := Article{
		ID:    "a",
		Title: "old",
		Kind:  ArticleKindTutorial,
		Level: ArticleLevelBeginner,
		Promo: true,
	}
	q := 1.4
	src.Quality = &q
	at := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	src.EnrichedAt = &at

	dst.CopySignals(src)
	if dst.Title != "new" || dst.Tags[0] != "go" {
		t.Fatalf("content changed: %#v", dst)
	}
	if dst.EnrichedAt == nil || !dst.EnrichedAt.Equal(at) {
		t.Fatalf("enriched_at: %#v", dst.EnrichedAt)
	}
	if dst.Kind != ArticleKindTutorial || dst.Level != ArticleLevelBeginner || !dst.Promo || dst.Quality == nil || *dst.Quality != 1.4 {
		t.Fatalf("signals: %#v", dst)
	}
}

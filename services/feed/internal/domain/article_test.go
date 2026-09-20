package domain

import "testing"

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

	dst.CopySignals(src)
	if dst.Title != "new" || dst.Tags[0] != "go" {
		t.Fatalf("content changed: %#v", dst)
	}
	if dst.Kind != ArticleKindTutorial || dst.Level != ArticleLevelBeginner || !dst.Promo || dst.Quality == nil || *dst.Quality != 1.4 {
		t.Fatalf("signals: %#v", dst)
	}
}

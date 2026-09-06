package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestBookmarkNormalized(t *testing.T) {
	got, err := Bookmark{
		UserID: "  u1  ",
		URL:    " https://zenn.dev/a/articles/hello ",
		Title:  "  Hello   Go  ",
		Source: " Zenn ",
	}.Normalized()
	if err != nil {
		t.Fatal(err)
	}
	if got.UserID != "u1" || got.Source != "zenn" || got.Title != "Hello Go" {
		t.Fatalf("%#v", got)
	}
	if got.ArticleID != IDFromURL("https://zenn.dev/a/articles/hello") {
		t.Fatalf("id: %q", got.ArticleID)
	}
}

func TestBookmarkNormalizedRejects(t *testing.T) {
	_, err := Bookmark{URL: "https://zenn.dev/x", Title: "t"}.Normalized()
	if !errors.Is(err, ErrUserIDRequired) {
		t.Fatalf("want user id, got %v", err)
	}
	_, err = Bookmark{UserID: "u1", URL: "ftp://x", Title: "t"}.Normalized()
	if !errors.Is(err, ErrInvalidArticleURL) {
		t.Fatalf("want url, got %v", err)
	}
	_, err = Bookmark{UserID: "u1", URL: "https://zenn.dev/x", Title: "   "}.Normalized()
	if !errors.Is(err, ErrArticleTitleRequired) {
		t.Fatalf("want title, got %v", err)
	}
	_, err = Bookmark{UserID: "u1", URL: "https://zenn.dev/x", Title: strings.Repeat("あ", MaxTitleRunes+1)}.Normalized()
	if !errors.Is(err, ErrArticleTitleRequired) {
		t.Fatalf("want long title, got %v", err)
	}
}

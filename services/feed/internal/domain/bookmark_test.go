package domain

import "testing"

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

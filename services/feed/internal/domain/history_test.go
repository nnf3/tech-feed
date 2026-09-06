package domain

import "testing"

func TestHistoryNormalized(t *testing.T) {
	got, err := HistoryEntry{
		UserID: " u1 ",
		URL:    "https://qiita.com/a/items/abc",
		Title:  "Hello",
		Source: "Qiita",
	}.Normalized()
	if err != nil {
		t.Fatal(err)
	}
	if got.Source != "qiita" || got.ViewCount != 1 || got.ArticleID == "" {
		t.Fatalf("%#v", got)
	}
}

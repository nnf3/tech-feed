package domain

import (
	"errors"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestNormalizeArticleSnapshot(t *testing.T) {
	url := "https://zenn.dev/a/articles/hello"
	got, err := normalizeArticleSnapshot("  u1  ", "", " "+url+" ", "  Hello   Go  ", " Zenn ")
	if err != nil {
		t.Fatal(err)
	}
	if got.userID != "u1" || got.title != "Hello Go" || got.source != "zenn" || got.url != url {
		t.Fatalf("%#v", got)
	}
	if got.articleID != IDFromURL(url) {
		t.Fatalf("id: %q", got.articleID)
	}
}

func TestNormalizeArticleSnapshotKeepsID(t *testing.T) {
	got, err := normalizeArticleSnapshot("u1", " given-id ", "https://qiita.com/a/items/x", "t", "qiita")
	if err != nil {
		t.Fatal(err)
	}
	if got.articleID != "given-id" {
		t.Fatalf("id: %q", got.articleID)
	}
}

func TestNormalizeArticleSnapshotRejects(t *testing.T) {
	cases := []struct {
		name string
		user string
		url  string
		title string
		want error
	}{
		{name: "no user", url: "https://zenn.dev/x", title: "t", want: ErrUserIDRequired},
		{name: "ftp", user: "u1", url: "ftp://x", title: "t", want: ErrInvalidArticleURL},
		{name: "no host", user: "u1", url: "https://", title: "t", want: ErrInvalidArticleURL},
		{name: "blank title", user: "u1", url: "https://zenn.dev/x", title: "   ", want: ErrArticleTitleRequired},
		{name: "long title", user: "u1", url: "https://zenn.dev/x", title: strings.Repeat("あ", MaxTitleRunes+1), want: ErrArticleTitleRequired},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := normalizeArticleSnapshot(tc.user, "", tc.url, tc.title, "zenn")
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v want %v", err, tc.want)
			}
		})
	}
}

func TestNormalizeArticleSnapshotTruncatesSource(t *testing.T) {
	got, err := normalizeArticleSnapshot("u1", "", "https://zenn.dev/x", "t", strings.Repeat("s", MaxSourceLen+8))
	if err != nil {
		t.Fatal(err)
	}
	if utf8.RuneCountInString(got.source) != MaxSourceLen {
		t.Fatalf("source len %d", utf8.RuneCountInString(got.source))
	}
}

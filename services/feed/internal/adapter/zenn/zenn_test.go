package zenn

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResourceFromURL(t *testing.T) {
	cases := []struct {
		raw  string
		kind string
		slug string
		ok   bool
	}{
		{"https://zenn.dev/gamella/articles/2b0d84fd597da3", "articles", "2b0d84fd597da3", true},
		{"https://zenn.dev/p/team/articles/abc123?query=1", "articles", "abc123", true},
		{"https://zenn.dev/saku0512/books/3735de8d0aa09f", "books", "3735de8d0aa09f", true},
		{"https://zenn.dev/feed", "", "", false},
	}
	for _, tc := range cases {
		kind, slug, ok := resourceFromURL(tc.raw)
		if kind != tc.kind || slug != tc.slug || ok != tc.ok {
			t.Fatalf("%s: got %s/%s ok=%v", tc.raw, kind, slug, ok)
		}
	}
}

func TestTopicsFromJSON(t *testing.T) {
	raw := []byte(`{"article":{"topics":[{"name":"Go"},{"name":"  rust  "},{"name":""}]}}`)
	got, err := topicsFromJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(got, ",") != "go,rust" {
		t.Fatalf("got %#v", got)
	}

	book, err := topicsFromJSON([]byte(`{"book":{"topics":[{"name":"C"}]}}`))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(book, ",") != "c" {
		t.Fatalf("book %#v", book)
	}
}

func TestFetchTopics(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/articles/2b0d84fd597da3" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"article":{"topics":[{"name":"クオンツ"},{"name":"トレーディング"}]}}`))
	}))
	t.Cleanup(srv.Close)

	src := New("")
	src.API = srv.URL
	src.client = srv.Client()

	got, err := src.fetchTopics(context.Background(), "articles", "2b0d84fd597da3")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(got, ",") != "クオンツ,トレーディング" {
		t.Fatalf("got %#v", got)
	}
}

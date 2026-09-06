package qiita

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[
			{"title":"Hello Go","url":"https://qiita.com/u/items/abc","body":"<p>hi</p>","created_at":"2026-09-06T00:00:00+09:00","tags":[{"name":"Go"},{"name":"API"}]},
			{"title":"","url":"https://qiita.com/u/items/skip","body":"","tags":[]}
		]`))
	}))
	t.Cleanup(srv.Close)

	src := New(srv.URL)
	src.client = srv.Client()
	got, err := src.Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("len=%d", len(got))
	}
	if got[0].Source != "qiita" || got[0].Title != "Hello Go" {
		t.Fatalf("article: %#v", got[0])
	}
	if strings.Join(got[0].Tags, ",") != "go,api" {
		t.Fatalf("tags: %#v", got[0].Tags)
	}
}

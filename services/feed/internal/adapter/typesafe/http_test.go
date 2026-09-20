package typesafe

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

func TestClientRetriesThenSucceeds(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`{"answers":{"kind":{"choice":"news","confidence":0.95},"promo":{"noul":0.0}}}`))
	}))
	t.Cleanup(srv.Close)

	client := New(Config{APIKey: "k", BaseURL: srv.URL, HTTPClient: srv.Client(), Concurrency: 1})
	got, err := client.Enrich(context.Background(), []domain.Article{{ID: "a1", Title: "n"}})
	if err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 || got[0].Kind != domain.ArticleKindNews {
		t.Fatalf("calls=%d got=%#v", calls.Load(), got)
	}
}

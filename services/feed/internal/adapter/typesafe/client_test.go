package typesafe

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

func TestClientEnrich(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/systemone" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("auth %q", got)
		}
		raw, _ := io.ReadAll(r.Body)
		var req map[string]any
		if err := json.Unmarshal(raw, &req); err != nil {
			t.Errorf("body: %v", err)
		}
		if req["model"] != "jev-latest" {
			t.Errorf("request %#v", req)
		}
		state, _ := req["state"].(map[string]any)
		if state["title"] != "Go GC" {
			t.Errorf("state %#v", state)
		}
		questions, _ := req["questions"].(map[string]any)
		if _, ok := questions[questionKind]; !ok {
			t.Errorf("questions %#v", questions)
		}
		calls.Add(1)
		_, _ = w.Write([]byte(`{
			"answers": {
				"kind": {"choice":"tutorial","confidence":0.9},
				"level": {"score":1,"confidence":0.8},
				"quality": {"score":1.7,"confidence":0.7},
				"promo": {"noul":0.1}
			}
		}`))
	}))
	t.Cleanup(srv.Close)

	client := New(Config{APIKey: "test-key", BaseURL: srv.URL, HTTPClient: srv.Client(), Concurrency: 1})
	got, err := client.Enrich(context.Background(), []domain.Article{
		{ID: "a1", Title: "Go GC", Summary: "internals", Tags: []string{"go"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 {
		t.Fatalf("calls %d", calls.Load())
	}
	if got[0].Kind != domain.ArticleKindTutorial || got[0].Level != domain.ArticleLevelIntermediate || got[0].Promo {
		t.Fatalf("%#v", got[0])
	}
	if got[0].Quality == nil || *got[0].Quality != 1.7 {
		t.Fatalf("quality %#v", got[0].Quality)
	}
	if got[0].EnrichedAt == nil {
		t.Fatalf("enriched_at not set: %#v", got[0])
	}
}

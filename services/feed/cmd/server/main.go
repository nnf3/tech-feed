package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nnf3/tech-feed/services/feed/internal/article"
	"github.com/nnf3/tech-feed/services/feed/internal/crawler"
	"github.com/nnf3/tech-feed/services/feed/internal/httpserver"
	"github.com/nnf3/tech-feed/services/feed/internal/search"
)

func main() {
	esHost := os.Getenv("ES_HOST")
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	store, err := search.New(esHost)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	if err := store.WaitReady(ctx); err != nil {
		cancel()
		log.Fatal(err)
	}
	if err := store.EnsureIndex(ctx); err != nil {
		cancel()
		log.Fatal(err)
	}
	cancel()

	zenn := crawler.NewZenn(os.Getenv("ZENN_FEED_URL"))
	ingest := func(ctx context.Context) ([]article.Article, error) {
		items, err := zenn.Fetch()
		if err != nil {
			return nil, err
		}
		if err := store.BulkUpsert(ctx, items); err != nil {
			return nil, err
		}
		log.Printf("ingested %d articles from zenn", len(items))
		return items, nil
	}

	go func() {
		ingestCtx, ingestCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer ingestCancel()
		if _, err := ingest(ingestCtx); err != nil {
			log.Printf("startup ingest skipped: %v", err)
		}
	}()

	srv := httpserver.New(store, ingest)
	log.Printf("feed listening on %s", addr)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}

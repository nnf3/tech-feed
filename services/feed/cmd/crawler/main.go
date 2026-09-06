package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nnf3/tech-feed/services/feed/internal/adapter/es"
	"github.com/nnf3/tech-feed/services/feed/internal/adapter/sources"
	"github.com/nnf3/tech-feed/services/feed/internal/usecase"
)

func main() {
	esHost := os.Getenv("ES_HOST")
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8081"
	}

	store, err := es.New(esHost)
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

	ingest := usecase.NewIngest(store, sources.All(os.Getenv("ZENN_FEED_URL"), os.Getenv("QIITA_API_URL"))...)
	crawler := usecase.NewCrawler(ingest, usecase.ParseInterval(os.Getenv("CRAWL_INTERVAL")), time.Minute)

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		})
		log.Printf("crawler health on %s", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatal(err)
		}
	}()

	crawler.Start(context.Background())
}

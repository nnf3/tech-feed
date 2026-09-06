package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nnf3/tech-feed/services/feed/internal/adapter/es"
	"github.com/nnf3/tech-feed/services/feed/internal/adapter/httpserver"
	"github.com/nnf3/tech-feed/services/feed/internal/adapter/ranking"
	"github.com/nnf3/tech-feed/services/feed/internal/adapter/zenn"
	"github.com/nnf3/tech-feed/services/feed/internal/usecase"
)

func main() {
	esHost := os.Getenv("ES_HOST")
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
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

	ingest := usecase.NewIngest(store, zenn.New(os.Getenv("ZENN_FEED_URL")))
	list := usecase.NewListFeed(store, ranking.PublishedAt{})

	// 起動時に一度だけ取り込む。定期クロールは crawler をサービスとして切り出すときに入れる。
	go func() {
		ingestCtx, ingestCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer ingestCancel()
		items, err := ingest.Run(ingestCtx)
		if err != nil {
			log.Printf("startup ingest skipped: %v", err)
			return
		}
		log.Printf("ingested %d articles from zenn", len(items))
	}()

	srv := httpserver.New(list, ingest)
	log.Printf("feed listening on %s", addr)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}

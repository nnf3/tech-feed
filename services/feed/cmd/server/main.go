package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nnf3/tech-feed/services/feed/internal/adapter/es"
	"github.com/nnf3/tech-feed/services/feed/internal/adapter/httpserver"
	"github.com/nnf3/tech-feed/services/feed/internal/adapter/postgres"
	"github.com/nnf3/tech-feed/services/feed/internal/adapter/ranking"
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
	defer cancel()
	if err := store.WaitReady(ctx); err != nil {
		log.Fatal(err)
	}
	if err := store.EnsureIndex(ctx); err != nil {
		log.Fatal(err)
	}

	db, err := postgres.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Migrate(ctx); err != nil {
		log.Fatal(err)
	}

	list := usecase.NewListFeed(store, db, ranking.PublishedAt{})
	users := usecase.NewUsers(db)
	profiles := usecase.NewProfiles(db, db)

	srv := httpserver.New(list, users, profiles)
	log.Printf("feed listening on %s", addr)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}

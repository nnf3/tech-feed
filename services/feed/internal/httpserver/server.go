package httpserver

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/nnf3/tech-feed/services/feed/internal/article"
	"github.com/nnf3/tech-feed/services/feed/internal/ranking"
	"github.com/nnf3/tech-feed/services/feed/internal/search"
)

type ingestFunc func(ctx context.Context) ([]article.Article, error)

type Server struct {
	store  *search.Store
	ingest ingestFunc
}

func New(store *search.Store, ingest ingestFunc) *Server {
	return &Server{store: store, ingest: ingest}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /articles", s.articles)
	mux.HandleFunc("POST /ingest", s.triggerIngest)
	return mux
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) articles(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	query := r.URL.Query().Get("q")
	items, err := s.store.Search(ctx, query)
	if err != nil {
		log.Printf("search: %v", err)
		http.Error(w, "search failed", http.StatusInternalServerError)
		return
	}
	if query == "" {
		items = ranking.ByPublishedAt(items)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"articles": items,
	})
}

func (s *Server) triggerIngest(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	items, err := s.ingest(ctx)
	if err != nil {
		log.Printf("ingest: %v", err)
		http.Error(w, "ingest failed", http.StatusBadGateway)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ingested": len(items)})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json: %v", err)
	}
}

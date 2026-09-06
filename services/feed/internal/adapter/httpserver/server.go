package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
	"github.com/nnf3/tech-feed/services/feed/internal/usecase"
)

type Server struct {
	list     *usecase.ListFeed
	ingest   *usecase.Ingest
	users    *usecase.Users
	profiles *usecase.Profiles
}

func New(list *usecase.ListFeed, ingest *usecase.Ingest, users *usecase.Users, profiles *usecase.Profiles) *Server {
	return &Server{list: list, ingest: ingest, users: users, profiles: profiles}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /articles", s.articles)
	mux.HandleFunc("POST /ingest", s.triggerIngest)
	mux.HandleFunc("PUT /users", s.upsertUser)
	mux.HandleFunc("GET /users/{id}", s.getUser)
	mux.HandleFunc("GET /users/{id}/profile", s.getProfile)
	mux.HandleFunc("PUT /users/{id}/profile", s.upsertProfile)
	return mux
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) articles(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	items, err := s.list.Run(ctx, r.URL.Query().Get("q"), r.URL.Query().Get("user_id"))
	if err != nil {
		log.Printf("list feed: %v", err)
		http.Error(w, "search failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"articles": items,
	})
}

func (s *Server) triggerIngest(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	items, err := s.ingest.Run(ctx)
	if err != nil {
		log.Printf("ingest: %v", err)
		http.Error(w, "ingest failed", http.StatusBadGateway)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ingested": len(items)})
}

func (s *Server) upsertUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var body domain.User
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	user, err := s.users.Upsert(ctx, body)
	if err != nil {
		writeUsecaseError(w, "upsert user", err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	user, err := s.users.Get(ctx, r.PathValue("id"))
	if err != nil {
		writeUsecaseError(w, "get user", err)
		return
	}
	if user == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (s *Server) getProfile(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	profile, err := s.profiles.Get(ctx, r.PathValue("id"))
	if err != nil {
		writeUsecaseError(w, "get profile", err)
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (s *Server) upsertProfile(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var body domain.Profile
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	body.UserID = r.PathValue("id")

	profile, err := s.profiles.Upsert(ctx, body)
	if err != nil {
		writeUsecaseError(w, "upsert profile", err)
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func writeUsecaseError(w http.ResponseWriter, op string, err error) {
	log.Printf("%s: %v", op, err)
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case domain.IsValidation(err):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json: %v", err)
	}
}

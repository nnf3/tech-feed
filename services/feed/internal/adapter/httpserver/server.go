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

type articleLister interface {
	Run(ctx context.Context, query, userID, tag string) ([]domain.Article, error)
}

type Server struct {
	list          articleLister
	users         *usecase.Users
	profiles      *usecase.Profiles
	internalToken string
}

func New(list *usecase.ListFeed, users *usecase.Users, profiles *usecase.Profiles, internalToken string) *Server {
	return &Server{list: list, users: users, profiles: profiles, internalToken: internalToken}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /articles", s.articles)
	mux.HandleFunc("GET /me", s.getUser)
	mux.HandleFunc("PUT /me", s.upsertUser)
	mux.HandleFunc("GET /me/profile", s.getProfile)
	mux.HandleFunc("PUT /me/profile", s.upsertProfile)
	return mux
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) articles(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	userID := ""
	if s.hasInternalToken(r) {
		userID = callerUserID(r)
	}

	items, err := s.list.Run(ctx, r.URL.Query().Get("q"), userID, r.URL.Query().Get("tag"))
	if err != nil {
		log.Printf("list feed: %v", err)
		http.Error(w, "search failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"articles": items,
	})
}

func (s *Server) upsertUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.identity(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var body domain.User
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	body.ID = userID

	user, err := s.users.Upsert(ctx, body)
	if err != nil {
		writeUsecaseError(w, "upsert user", err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.identity(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	user, err := s.users.Get(ctx, userID)
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
	userID, ok := s.identity(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	profile, err := s.profiles.Get(ctx, userID)
	if err != nil {
		writeUsecaseError(w, "get profile", err)
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (s *Server) upsertProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.identity(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var body domain.Profile
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	body.UserID = userID

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

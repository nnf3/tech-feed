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
	Run(ctx context.Context, query, userID, tag, sort, after string) (domain.FeedPage, error)
}

type Server struct {
	list          articleLister
	users         *usecase.Users
	profiles      *usecase.Profiles
	bookmarks     *usecase.Bookmarks
	history       *usecase.History
	internalToken string
}

func New(list *usecase.ListFeed, users *usecase.Users, profiles *usecase.Profiles, bookmarks *usecase.Bookmarks, history *usecase.History, internalToken string) *Server {
	return &Server{list: list, users: users, profiles: profiles, bookmarks: bookmarks, history: history, internalToken: internalToken}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /articles", s.articles)
	mux.HandleFunc("GET /me", s.getUser)
	mux.HandleFunc("PUT /me", s.upsertUser)
	mux.HandleFunc("GET /me/profile", s.getProfile)
	mux.HandleFunc("PUT /me/profile", s.upsertProfile)
	mux.HandleFunc("GET /me/bookmarks", s.listBookmarks)
	mux.HandleFunc("PUT /me/bookmarks", s.addBookmark)
	mux.HandleFunc("DELETE /me/bookmarks/{article_id}", s.removeBookmark)
	mux.HandleFunc("GET /me/history", s.listHistory)
	mux.HandleFunc("PUT /me/history", s.recordHistory)
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

	page, err := s.list.Run(
		ctx,
		r.URL.Query().Get("q"),
		userID,
		r.URL.Query().Get("tag"),
		r.URL.Query().Get("sort"),
		r.URL.Query().Get("after"),
	)
	if err != nil {
		log.Printf("list feed: %v", err)
		if domain.IsValidation(err) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "search failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, page)
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

func (s *Server) listBookmarks(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.identity(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	items, err := s.bookmarks.List(ctx, userID)
	if err != nil {
		writeUsecaseError(w, "list bookmarks", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"bookmarks": items})
}

func (s *Server) addBookmark(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.identity(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var body domain.Bookmark
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	body.UserID = userID

	item, err := s.bookmarks.Add(ctx, body)
	if err != nil {
		writeUsecaseError(w, "add bookmark", err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) removeBookmark(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.identity(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := s.bookmarks.Remove(ctx, userID, r.PathValue("article_id")); err != nil {
		writeUsecaseError(w, "remove bookmark", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) listHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.identity(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	items, err := s.history.List(ctx, userID)
	if err != nil {
		writeUsecaseError(w, "list history", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"history": items})
}

func (s *Server) recordHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.identity(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var body domain.HistoryEntry
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	body.UserID = userID

	item, err := s.history.Record(ctx, body)
	if err != nil {
		writeUsecaseError(w, "record history", err)
		return
	}
	writeJSON(w, http.StatusOK, item)
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

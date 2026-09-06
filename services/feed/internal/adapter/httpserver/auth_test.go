package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

type stubList struct {
	userID string
}

func (s *stubList) Run(_ context.Context, _, userID, _, _, _ string) (domain.FeedPage, error) {
	s.userID = userID
	return domain.FeedPage{Articles: []domain.Article{}}, nil
}

func testServer(token string, list articleLister) http.Handler {
	return (&Server{list: list, internalToken: token}).Handler()
}

func TestUsersRequireInternalToken(t *testing.T) {
	h := testServer("secret", nil)
	cases := []struct {
		name string
		auth string
		user string
		want int
	}{
		{name: "none", want: http.StatusUnauthorized},
		{name: "bad token", auth: "Bearer nope", user: "alice", want: http.StatusUnauthorized},
		{name: "no user", auth: "Bearer secret", want: http.StatusUnauthorized},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/me", nil)
			if tc.auth != "" {
				req.Header.Set("Authorization", tc.auth)
			}
			if tc.user != "" {
				req.Header.Set("X-User-ID", tc.user)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("got %d want %d", rec.Code, tc.want)
			}
		})
	}
}

func TestHistoryRequireInternalToken(t *testing.T) {
	h := testServer("secret", nil)
	req := httptest.NewRequest(http.MethodGet, "/me/history", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestBookmarksRequireInternalToken(t *testing.T) {
	h := testServer("secret", nil)
	req := httptest.NewRequest(http.MethodGet, "/me/bookmarks", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestOldUsersPathIsGone(t *testing.T) {
	h := testServer("secret", nil)
	req := httptest.NewRequest(http.MethodGet, "/users/alice", nil)
	req.Header.Set("Authorization", "Bearer secret")
	req.Header.Set("X-User-ID", "alice")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestArticlesIgnoreQueryUserID(t *testing.T) {
	list := &stubList{}
	h := testServer("secret", list)

	req := httptest.NewRequest(http.MethodGet, "/articles?user_id=alice", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d", rec.Code)
	}
	if list.userID != "" {
		t.Fatalf("personalized without token: %q", list.userID)
	}
}

func TestArticlesPersonalizeWithToken(t *testing.T) {
	list := &stubList{}
	h := testServer("secret", list)

	req := httptest.NewRequest(http.MethodGet, "/articles?user_id=eve", nil)
	req.Header.Set("Authorization", "Bearer secret")
	req.Header.Set("X-User-ID", "alice")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d", rec.Code)
	}
	if list.userID != "alice" {
		t.Fatalf("got user %q", list.userID)
	}
}

func TestEqualSecret(t *testing.T) {
	if equalSecret("a", "") || equalSecret("ab", "abc") || !equalSecret("tok", "tok") {
		t.Fatal("equalSecret mismatch")
	}
}

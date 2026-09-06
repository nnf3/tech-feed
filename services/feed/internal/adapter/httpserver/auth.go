package httpserver

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

const headerUserID = "X-User-ID"

func bearerToken(raw string) string {
	const prefix = "Bearer "
	if len(raw) < len(prefix) || !strings.EqualFold(raw[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(raw[len(prefix):])
}

func equalSecret(got, want string) bool {
	if want == "" {
		return false
	}
	a := []byte(got)
	b := []byte(want)
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare(a, b) == 1
}

func (s *Server) hasInternalToken(r *http.Request) bool {
	return equalSecret(bearerToken(r.Header.Get("Authorization")), s.internalToken)
}

func callerUserID(r *http.Request) string {
	return strings.TrimSpace(r.Header.Get(headerUserID))
}

// identity は BFF の共有トークンと X-User-ID を見る。パスにユーザー ID は載せない。
func (s *Server) identity(w http.ResponseWriter, r *http.Request) (string, bool) {
	if !s.hasInternalToken(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return "", false
	}
	userID := callerUserID(r)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return "", false
	}
	return userID, true
}

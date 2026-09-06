package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestProfileNormalized(t *testing.T) {
	got, err := Profile{
		UserID:       "  u1  ",
		InterestTags: []string{" Go ", "go", "RUST"},
		ExcludeTags:  []string{"beginner", ""},
	}.Normalized()
	if err != nil {
		t.Fatal(err)
	}
	if got.UserID != "u1" {
		t.Fatalf("user id: %q", got.UserID)
	}
	if strings.Join(got.InterestTags, ",") != "go,rust" {
		t.Fatalf("interest: %#v", got.InterestTags)
	}
	if strings.Join(got.ExcludeTags, ",") != "beginner" {
		t.Fatalf("exclude: %#v", got.ExcludeTags)
	}
}

func TestProfileNormalizedOverlap(t *testing.T) {
	_, err := Profile{UserID: "u1", InterestTags: []string{"go"}, ExcludeTags: []string{"GO"}}.Normalized()
	if !errors.Is(err, ErrTagOverlap) {
		t.Fatalf("want overlap, got %v", err)
	}
}

func TestUserNormalizedRequiresID(t *testing.T) {
	_, err := User{Email: "a@b.c"}.Normalized()
	if !errors.Is(err, ErrUserIDRequired) {
		t.Fatalf("want id required, got %v", err)
	}
}

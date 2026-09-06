package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

type stubIndex struct {
	last domain.FeedQuery
}

func (s *stubIndex) BulkUpsert(context.Context, []domain.Article) error { return nil }

func (s *stubIndex) Search(_ context.Context, query domain.FeedQuery) (domain.FeedPage, error) {
	s.last = query
	return domain.FeedPage{Articles: []domain.Article{{ID: "1"}}}, nil
}

type stubProfiles struct {
	profile *domain.Profile
}

func (s stubProfiles) UpsertProfile(context.Context, domain.Profile) error { return nil }

func (s stubProfiles) GetProfile(context.Context, string) (*domain.Profile, error) {
	return s.profile, nil
}

func TestListFeedRecommendedUsesInterest(t *testing.T) {
	index := &stubIndex{}
	u := NewListFeed(index, stubProfiles{profile: &domain.Profile{
		InterestTags: []string{"go"},
		ExcludeTags:  []string{"poem"},
	}})

	if _, err := u.Run(context.Background(), "", "u1", "", domain.SortRecommended, ""); err != nil {
		t.Fatal(err)
	}
	if len(index.last.InterestTags) != 1 || index.last.Sort != domain.SortRecommended {
		t.Fatalf("%#v", index.last)
	}
	if len(index.last.ExcludeTags) != 1 {
		t.Fatalf("exclude: %#v", index.last.ExcludeTags)
	}

	if _, err := u.Run(context.Background(), "", "u1", "", domain.SortNew, ""); err != nil {
		t.Fatal(err)
	}
	if len(index.last.InterestTags) != 0 || index.last.Sort != domain.SortNew {
		t.Fatalf("new should drop interest: %#v", index.last)
	}
}

func TestListFeedRejectsBadAfter(t *testing.T) {
	u := NewListFeed(&stubIndex{}, stubProfiles{})
	_, err := u.Run(context.Background(), "", "", "", "", "%%%")
	if !errors.Is(err, domain.ErrInvalidSearchAfter) {
		t.Fatalf("got %v", err)
	}
}

func TestListFeedGuestCannotRecommend(t *testing.T) {
	index := &stubIndex{}
	u := NewListFeed(index, stubProfiles{})
	if _, err := u.Run(context.Background(), "", "", "", domain.SortRecommended, ""); err != nil {
		t.Fatal(err)
	}
	if index.last.Sort != domain.SortNew {
		t.Fatalf("guest sort: %q", index.last.Sort)
	}
}

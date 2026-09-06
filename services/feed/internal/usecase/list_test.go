package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

type stubIndex struct {
	last domain.FeedQuery
	docs map[string]domain.Article
}

func (s *stubIndex) BulkUpsert(context.Context, []domain.Article) error { return nil }

func (s *stubIndex) GetByIDs(_ context.Context, ids []string) ([]domain.Article, error) {
	var out []domain.Article
	for _, id := range ids {
		if article, ok := s.docs[id]; ok {
			out = append(out, article)
		}
	}
	return out, nil
}

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
	}}, nil, nil)

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
	u := NewListFeed(&stubIndex{}, stubProfiles{}, nil, nil)
	_, err := u.Run(context.Background(), "", "", "", "", "%%%")
	if !errors.Is(err, domain.ErrInvalidSearchAfter) {
		t.Fatalf("got %v", err)
	}
}

func TestListFeedGuestCannotRecommend(t *testing.T) {
	index := &stubIndex{}
	u := NewListFeed(index, stubProfiles{}, nil, nil)
	if _, err := u.Run(context.Background(), "", "", "", domain.SortRecommended, ""); err != nil {
		t.Fatal(err)
	}
	if index.last.Sort != domain.SortNew {
		t.Fatalf("guest sort: %q", index.last.Sort)
	}
}

type stubBookmarks struct {
	items []domain.Bookmark
}

func (s stubBookmarks) UpsertBookmark(context.Context, domain.Bookmark) error { return nil }
func (s stubBookmarks) GetBookmark(context.Context, string, string) (*domain.Bookmark, error) {
	return nil, nil
}
func (s stubBookmarks) DeleteBookmark(context.Context, string, string) error { return nil }
func (s stubBookmarks) ListBookmarks(context.Context, string) ([]domain.Bookmark, error) {
	return s.items, nil
}
func (s stubBookmarks) CountBookmarks(context.Context, string) (int, error) { return len(s.items), nil }

type stubHistory struct {
	items []domain.HistoryEntry
}

func (s stubHistory) UpsertHistory(context.Context, domain.HistoryEntry) error { return nil }
func (s stubHistory) GetHistory(context.Context, string, string) (*domain.HistoryEntry, error) {
	return nil, nil
}
func (s stubHistory) ListHistory(context.Context, string) ([]domain.HistoryEntry, error) {
	return s.items, nil
}
func (s stubHistory) TrimHistory(context.Context, string, int) error { return nil }

func TestListFeedRecommendedUsesActivity(t *testing.T) {
	index := &stubIndex{docs: map[string]domain.Article{
		"b1": {ID: "b1", Tags: []string{"kubernetes", "poem"}},
		"h1": {ID: "h1", Tags: []string{"rust", "cli"}},
	}}
	u := NewListFeed(
		index,
		stubProfiles{profile: &domain.Profile{ExcludeTags: []string{"poem"}}},
		stubBookmarks{items: []domain.Bookmark{{ArticleID: "b1"}}},
		stubHistory{items: []domain.HistoryEntry{{ArticleID: "h1", ViewCount: 2}}},
	)

	if _, err := u.Run(context.Background(), "", "u1", "", domain.SortRecommended, ""); err != nil {
		t.Fatal(err)
	}
	if len(index.last.BookmarkTags) != 1 || index.last.BookmarkTags[0] != "kubernetes" {
		t.Fatalf("bookmark tags: %#v", index.last.BookmarkTags)
	}
	if len(index.last.HistoryTags) != 2 {
		t.Fatalf("history tags: %#v", index.last.HistoryTags)
	}
	if len(index.last.ExcludeIDs) != 2 {
		t.Fatalf("exclude ids: %#v", index.last.ExcludeIDs)
	}

	if _, err := u.Run(context.Background(), "", "u1", "", domain.SortNew, ""); err != nil {
		t.Fatal(err)
	}
	if len(index.last.BookmarkTags) != 0 || len(index.last.ExcludeIDs) != 0 {
		t.Fatalf("new leaked activity: %#v", index.last)
	}
}

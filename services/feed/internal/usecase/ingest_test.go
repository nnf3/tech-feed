package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

type stubSource struct {
	items []domain.Article
	err   error
}

func (s stubSource) Fetch(context.Context) ([]domain.Article, error) {
	return s.items, s.err
}

type ingestIndex struct {
	docs   map[string]domain.Article
	upsert []domain.Article
}

func (s *ingestIndex) BulkUpsert(_ context.Context, articles []domain.Article) error {
	s.upsert = append([]domain.Article(nil), articles...)
	return nil
}

func (s *ingestIndex) GetByIDs(_ context.Context, ids []string) ([]domain.Article, error) {
	var out []domain.Article
	for _, id := range ids {
		if article, ok := s.docs[id]; ok {
			out = append(out, article)
		}
	}
	return out, nil
}

func (s *ingestIndex) Search(context.Context, domain.FeedQuery) (domain.FeedPage, error) {
	return domain.FeedPage{}, nil
}

type stubEnricher struct {
	kind string
	err  error
	seen *[]string
}

func (s stubEnricher) Enrich(_ context.Context, articles []domain.Article) ([]domain.Article, error) {
	out := append([]domain.Article(nil), articles...)
	for i := range out {
		if s.seen != nil {
			*s.seen = append(*s.seen, out[i].ID)
		}
		if s.kind != "" {
			out[i].Kind = s.kind
			out[i].Promo = false
		}
	}
	return out, s.err
}

func TestIngestRestoresAndEnrichesOnlyPending(t *testing.T) {
	q := 0.2
	at := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	index := &ingestIndex{docs: map[string]domain.Article{
		"a1": {ID: "a1", Kind: domain.ArticleKindNews, Level: domain.ArticleLevelAdvanced, Quality: &q, EnrichedAt: &at},
	}}
	var seen []string
	u := NewIngest(index, stubEnricher{kind: domain.ArticleKindTutorial, seen: &seen}, stubSource{items: []domain.Article{
		{ID: "a1", Title: "fresh"},
		{ID: "a2", Title: "new"},
	}})

	got, err := u.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(seen) != 1 || seen[0] != "a2" {
		t.Fatalf("enriched ids: %v", seen)
	}
	if len(got) != 2 || got[0].Title != "fresh" || got[1].Title != "new" {
		t.Fatalf("content: %#v", got)
	}
	if got[0].Kind != domain.ArticleKindNews || got[0].Level != domain.ArticleLevelAdvanced || got[0].Quality == nil || *got[0].Quality != 0.2 {
		t.Fatalf("kept previous signals: %#v", got[0])
	}
	if got[0].EnrichedAt == nil || !got[0].EnrichedAt.Equal(at) {
		t.Fatalf("enriched_at: %#v", got[0].EnrichedAt)
	}
	if got[1].Kind != domain.ArticleKindTutorial {
		t.Fatalf("new article kind: %#v", got[1])
	}
	if len(index.upsert) != 2 || index.upsert[1].Kind != domain.ArticleKindTutorial {
		t.Fatalf("upsert: %#v", index.upsert)
	}
}

func TestIngestSkipsLegacySignalsWithoutEnrichedAt(t *testing.T) {
	index := &ingestIndex{docs: map[string]domain.Article{
		"a1": {ID: "a1", Kind: domain.ArticleKindOpinion},
	}}
	var seen []string
	u := NewIngest(index, stubEnricher{kind: domain.ArticleKindTutorial, seen: &seen}, stubSource{items: []domain.Article{
		{ID: "a1", Title: "fresh"},
	}})

	got, err := u.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(seen) != 0 {
		t.Fatalf("legacy article re-enriched: %v", seen)
	}
	if got[0].Kind != domain.ArticleKindOpinion {
		t.Fatalf("wanted previous signals: %#v", got[0])
	}
}

func TestIngestKeepsArticlesWhenEnrichFails(t *testing.T) {
	index := &ingestIndex{}
	u := NewIngest(index, failingEnricher{}, stubSource{items: []domain.Article{
		{ID: "a1", Title: "fresh"},
	}})

	got, err := u.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Title != "fresh" || got[0].Kind != "" {
		t.Fatalf("wanted original article: %#v", got)
	}
}

type failingEnricher struct{}

func (failingEnricher) Enrich(context.Context, []domain.Article) ([]domain.Article, error) {
	return nil, context.DeadlineExceeded
}

func TestIngestUsesPartialEnrichmentOnError(t *testing.T) {
	index := &ingestIndex{}
	u := NewIngest(index, stubEnricher{kind: domain.ArticleKindTutorial, err: context.Canceled}, stubSource{items: []domain.Article{
		{ID: "a1", Title: "fresh"},
	}})
	got, err := u.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Kind != domain.ArticleKindTutorial {
		t.Fatalf("partial enrichment dropped: %#v", got[0])
	}
}

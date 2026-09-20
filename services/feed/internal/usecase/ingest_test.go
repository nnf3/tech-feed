package usecase

import (
	"context"
	"testing"

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
}

func (s stubEnricher) Enrich(_ context.Context, articles []domain.Article) ([]domain.Article, error) {
	out := append([]domain.Article(nil), articles...)
	if s.kind != "" {
		for i := range out {
			out[i].Kind = s.kind
			out[i].Promo = false
		}
	}
	return out, s.err
}

func TestIngestRestoresThenEnriches(t *testing.T) {
	q := 0.2
	index := &ingestIndex{docs: map[string]domain.Article{
		"a1": {ID: "a1", Kind: domain.ArticleKindNews, Level: domain.ArticleLevelAdvanced, Quality: &q, Promo: true},
	}}
	u := NewIngest(index, stubEnricher{kind: domain.ArticleKindTutorial}, stubSource{items: []domain.Article{
		{ID: "a1", Title: "fresh"},
	}})

	got, err := u.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Title != "fresh" {
		t.Fatalf("content: %#v", got)
	}
	if got[0].Kind != domain.ArticleKindTutorial {
		t.Fatalf("kind: %#v", got[0])
	}
	if got[0].Level != domain.ArticleLevelAdvanced || got[0].Quality == nil || *got[0].Quality != 0.2 {
		t.Fatalf("kept previous signals: %#v", got[0])
	}
	if len(index.upsert) != 1 || index.upsert[0].Kind != domain.ArticleKindTutorial {
		t.Fatalf("upsert: %#v", index.upsert)
	}
}

func TestIngestKeepsSignalsWhenEnrichFails(t *testing.T) {
	index := &ingestIndex{docs: map[string]domain.Article{
		"a1": {ID: "a1", Kind: domain.ArticleKindOpinion, Promo: true},
	}}
	u := NewIngest(index, stubEnricher{err: context.DeadlineExceeded}, stubSource{items: []domain.Article{
		{ID: "a1", Title: "fresh"},
	}})

	got, err := u.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Kind != domain.ArticleKindOpinion || !got[0].Promo {
		t.Fatalf("wanted previous signals: %#v", got[0])
	}
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

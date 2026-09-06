package usecase

import (
	"context"
	"fmt"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

type Ingest struct {
	sources []domain.Source
	index   domain.ArticleIndex
}

func NewIngest(index domain.ArticleIndex, sources ...domain.Source) *Ingest {
	return &Ingest{index: index, sources: sources}
}

func (u *Ingest) Run(ctx context.Context) ([]domain.Article, error) {
	var all []domain.Article
	for _, src := range u.sources {
		items, err := src.Fetch(ctx)
		if err != nil {
			return nil, fmt.Errorf("fetch source: %w", err)
		}
		all = append(all, items...)
	}
	if err := u.index.BulkUpsert(ctx, all); err != nil {
		return nil, fmt.Errorf("upsert articles: %w", err)
	}
	return all, nil
}

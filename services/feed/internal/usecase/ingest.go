package usecase

import (
	"context"
	"fmt"
	"log"

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
	var lastErr error
	for _, src := range u.sources {
		items, err := src.Fetch(ctx)
		if err != nil {
			log.Printf("fetch source skipped: %v", err)
			lastErr = err
			continue
		}
		all = append(all, items...)
	}
	if len(all) == 0 {
		if lastErr != nil {
			return nil, fmt.Errorf("fetch sources: %w", lastErr)
		}
		return all, nil
	}
	if err := u.index.BulkUpsert(ctx, all); err != nil {
		return nil, fmt.Errorf("upsert articles: %w", err)
	}
	return all, nil
}

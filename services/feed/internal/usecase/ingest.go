package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

type Ingest struct {
	sources  []domain.Source
	index    domain.ArticleIndex
	enricher domain.ArticleEnricher
}

func NewIngest(index domain.ArticleIndex, enricher domain.ArticleEnricher, sources ...domain.Source) *Ingest {
	if enricher == nil {
		enricher = domain.NoopArticleEnricher()
	}
	return &Ingest{index: index, enricher: enricher, sources: sources}
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

	all = u.restoreSignals(ctx, all)
	enriched, err := u.enricher.Enrich(ctx, all)
	if err != nil {
		log.Printf("enrich articles skipped: %v", err)
	}
	if len(enriched) > 0 {
		all = enriched
	}

	if err := u.index.BulkUpsert(ctx, all); err != nil {
		return nil, fmt.Errorf("upsert articles: %w", err)
	}
	return all, nil
}

// restoreSignals は再クロールで判定が落ちても、前回の kind / level などを残す。
func (u *Ingest) restoreSignals(ctx context.Context, articles []domain.Article) []domain.Article {
	ids := make([]string, 0, len(articles))
	for _, article := range articles {
		ids = append(ids, article.ID)
	}
	previous, err := u.index.GetByIDs(ctx, ids)
	if err != nil {
		log.Printf("load article signals skipped: %v", err)
		return articles
	}
	byID := domain.ArticlesByID(previous)
	for i := range articles {
		if old, ok := byID[articles[i].ID]; ok {
			articles[i].CopySignals(old)
		}
	}
	return articles
}

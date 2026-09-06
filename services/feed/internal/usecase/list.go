package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

type ListFeed struct {
	index  domain.ArticleIndex
	ranker domain.Ranker
}

func NewListFeed(index domain.ArticleIndex, ranker domain.Ranker) *ListFeed {
	return &ListFeed{index: index, ranker: ranker}
}

func (u *ListFeed) Run(ctx context.Context, query string) ([]domain.Article, error) {
	items, err := u.index.Search(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("search articles: %w", err)
	}
	if strings.TrimSpace(query) == "" {
		return u.ranker.Rank(items), nil
	}
	return items, nil
}

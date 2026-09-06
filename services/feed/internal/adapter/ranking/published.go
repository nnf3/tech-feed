package ranking

import (
	"sort"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

// PublishedAt は現在の Ranker。後から recommendation クライアントに差し替える。
type PublishedAt struct{}

func (PublishedAt) Rank(articles []domain.Article) []domain.Article {
	sorted := append([]domain.Article(nil), articles...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].PublishedAt.After(sorted[j].PublishedAt)
	})
	return sorted
}

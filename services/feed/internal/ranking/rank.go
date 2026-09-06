package ranking

import (
	"sort"

	"github.com/nnf3/tech-feed/services/feed/internal/article"
)

// ByPublishedAt returns articles newest-first.
// Later this is where profile tags and embeddings will rerank.
func ByPublishedAt(articles []article.Article) []article.Article {
	sorted := append([]article.Article(nil), articles...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].PublishedAt.After(sorted[j].PublishedAt)
	})
	return sorted
}

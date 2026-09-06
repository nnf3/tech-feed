package sources

import (
	"github.com/nnf3/tech-feed/services/feed/internal/adapter/qiita"
	"github.com/nnf3/tech-feed/services/feed/internal/adapter/zenn"
	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

// All は取り込み対象のソース。追加するときはここだけ変える。
func All(zennURL, qiitaURL string) []domain.Source {
	return []domain.Source{
		zenn.New(zennURL),
		qiita.New(qiitaURL),
	}
}

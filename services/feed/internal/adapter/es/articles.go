package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

const articlesIndex = "articles"

// articlesMapping は articles インデックスの定義。
// 変更時は新規インデックス作成 + alias swap が必要（未実装）。
const articlesMapping = `{
  "settings": {
    "analysis": {
      "analyzer": {
        "ja_analyzer": {
          "type": "custom",
          "tokenizer": "kuromoji_tokenizer",
          "filter": ["kuromoji_baseform", "lowercase", "icu_normalizer"]
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "source": { "type": "keyword" },
      "url": { "type": "keyword" },
      "title": { "type": "text", "analyzer": "ja_analyzer" },
      "summary": { "type": "text", "analyzer": "ja_analyzer" },
      "tags": { "type": "keyword" },
      "published_at": { "type": "date" }
    }
  }
}`

func (s *Store) EnsureIndex(ctx context.Context) error {
	res, err := s.client.Indices.Exists([]string{articlesIndex}, s.client.Indices.Exists.WithContext(ctx))
	if err != nil {
		return err
	}
	res.Body.Close()
	if res.StatusCode == 200 {
		return nil
	}

	create, err := s.client.Indices.Create(
		articlesIndex,
		s.client.Indices.Create.WithContext(ctx),
		s.client.Indices.Create.WithBody(strings.NewReader(articlesMapping)),
	)
	if err != nil {
		return err
	}
	defer create.Body.Close()
	if create.IsError() {
		body, _ := io.ReadAll(create.Body)
		return fmt.Errorf("create index: %s", body)
	}
	return nil
}

func (s *Store) BulkUpsert(ctx context.Context, articles []domain.Article) error {
	if len(articles) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for _, a := range articles {
		meta, err := json.Marshal(map[string]any{
			"index": map[string]any{"_index": articlesIndex, "_id": a.ID},
		})
		if err != nil {
			return err
		}
		doc, err := json.Marshal(a)
		if err != nil {
			return err
		}
		buf.Write(meta)
		buf.WriteByte('\n')
		buf.Write(doc)
		buf.WriteByte('\n')
	}

	res, err := s.client.Bulk(bytes.NewReader(buf.Bytes()), s.client.Bulk.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("bulk index: %s", body)
	}
	return nil
}

type articleHit struct {
	Source domain.Article `json:"_source"`
}

type articleSearchResponse struct {
	Hits struct {
		Hits []articleHit `json:"hits"`
	} `json:"hits"`
}

func articleSearchBody(query string) map[string]any {
	if strings.TrimSpace(query) == "" {
		return map[string]any{
			"size": 50,
			"sort": []any{map[string]any{"published_at": map[string]any{"order": "desc"}}},
			"query": map[string]any{
				"match_all": map[string]any{},
			},
		}
	}
	return map[string]any{
		"size": 50,
		"sort": []any{
			map[string]any{"_score": map[string]any{"order": "desc"}},
			map[string]any{"published_at": map[string]any{"order": "desc"}},
		},
		"query": map[string]any{
			"multi_match": map[string]any{
				"query":    query,
				"fields":   []string{"title^2", "summary"},
				"analyzer": "ja_analyzer",
			},
		},
	}
}

func (s *Store) Search(ctx context.Context, query string) ([]domain.Article, error) {
	payload, err := json.Marshal(articleSearchBody(query))
	if err != nil {
		return nil, err
	}

	res, err := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(articlesIndex),
		s.client.Search.WithBody(bytes.NewReader(payload)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		raw, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("search: %s", raw)
	}

	var parsed articleSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	articles := make([]domain.Article, 0, len(parsed.Hits.Hits))
	for _, hit := range parsed.Hits.Hits {
		articles = append(articles, hit.Source)
	}
	return articles, nil
}

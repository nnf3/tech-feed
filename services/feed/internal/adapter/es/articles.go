package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

const (
	articlesAlias   = "articles"
	articlesVersion = 1
)

// articlesMapping は実インデックスの定義。
// 変えるときは articlesVersion を上げる。起動時に新インデックスへ reindex して alias を張り替える。
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

func articlesPhysicalIndex() string {
	return fmt.Sprintf("%s-v%d", articlesAlias, articlesVersion)
}

func (s *Store) EnsureIndex(ctx context.Context) error {
	desired := articlesPhysicalIndex()

	exists, err := s.existsIndex(ctx, desired)
	if err != nil {
		return err
	}
	if !exists {
		log.Printf("elasticsearch: create index %s", desired)
		if err := s.createIndex(ctx, desired, articlesMapping); err != nil {
			return err
		}
	}

	isAlias, err := s.existsAlias(ctx, articlesAlias)
	if err != nil {
		return err
	}
	if !isAlias {
		// 旧実装は alias と同名の実インデックスを作っていた。名前を空けるため、ここだけ一瞬 404 になり得る。
		legacy, err := s.existsIndex(ctx, articlesAlias)
		if err != nil {
			return err
		}
		if legacy {
			log.Printf("elasticsearch: migrate concrete index %s -> %s", articlesAlias, desired)
			if err := s.reindex(ctx, articlesAlias, desired); err != nil {
				return err
			}
			if err := s.deleteIndex(ctx, articlesAlias); err != nil {
				return err
			}
		}
		log.Printf("elasticsearch: point alias %s -> %s", articlesAlias, desired)
		return s.swapAlias(ctx, articlesAlias, desired, nil)
	}

	current, err := s.aliasTargets(ctx, articlesAlias)
	if err != nil {
		return err
	}
	if len(current) == 1 && current[0] == desired {
		return nil
	}

	log.Printf("elasticsearch: swap alias %s from %v to %s", articlesAlias, current, desired)
	for _, old := range current {
		if old == desired {
			continue
		}
		if err := s.reindex(ctx, old, desired); err != nil {
			return err
		}
	}
	if err := s.swapAlias(ctx, articlesAlias, desired, current); err != nil {
		return err
	}
	for _, old := range current {
		if old == desired {
			continue
		}
		if err := s.deleteIndex(ctx, old); err != nil {
			return err
		}
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
			"index": map[string]any{"_index": articlesAlias, "_id": a.ID},
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
		s.client.Search.WithIndex(articlesAlias),
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

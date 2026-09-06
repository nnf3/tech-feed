package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	elasticsearch "github.com/elastic/go-elasticsearch/v9"
	"github.com/nnf3/tech-feed/services/feed/internal/article"
)

const indexName = "articles"

type Store struct {
	client *elasticsearch.Client
}

func New(esHost string) (*Store, error) {
	if esHost == "" {
		esHost = "http://localhost:9200"
	}
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{esHost},
	})
	if err != nil {
		return nil, fmt.Errorf("create elasticsearch client: %w", err)
	}
	return &Store{client: client}, nil
}

func (s *Store) WaitReady(ctx context.Context) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		res, err := s.client.Info(s.client.Info.WithContext(ctx))
		if err == nil {
			res.Body.Close()
			if res.StatusCode < 300 {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			if err != nil {
				return fmt.Errorf("elasticsearch not ready: %w", err)
			}
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (s *Store) EnsureIndex(ctx context.Context) error {
	res, err := s.client.Indices.Exists([]string{indexName}, s.client.Indices.Exists.WithContext(ctx))
	if err != nil {
		return err
	}
	res.Body.Close()
	if res.StatusCode == 200 {
		return nil
	}

	mapping := `{
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

	create, err := s.client.Indices.Create(
		indexName,
		s.client.Indices.Create.WithContext(ctx),
		s.client.Indices.Create.WithBody(strings.NewReader(mapping)),
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

func (s *Store) BulkUpsert(ctx context.Context, articles []article.Article) error {
	if len(articles) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for _, a := range articles {
		meta, err := json.Marshal(map[string]any{
			"index": map[string]any{"_index": indexName, "_id": a.ID},
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

type esHit struct {
	Source article.Article `json:"_source"`
}

type esSearchResponse struct {
	Hits struct {
		Hits []esHit `json:"hits"`
	} `json:"hits"`
}

func (s *Store) Search(ctx context.Context, query string) ([]article.Article, error) {
	var body map[string]any
	if strings.TrimSpace(query) == "" {
		body = map[string]any{
			"size": 50,
			"sort": []any{map[string]any{"published_at": map[string]any{"order": "desc"}}},
			"query": map[string]any{
				"match_all": map[string]any{},
			},
		}
	} else {
		body = map[string]any{
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

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	res, err := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(indexName),
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

	var parsed esSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	articles := make([]article.Article, 0, len(parsed.Hits.Hits))
	for _, hit := range parsed.Hits.Hits {
		articles = append(articles, hit.Source)
	}
	return articles, nil
}

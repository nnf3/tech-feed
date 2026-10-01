package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

// BulkUpsert は記事を alias へ一括 index する。同じ ID は上書き。
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

// GetByIDs は ID 指定で記事を取る。おすすめ用に tags だけ読む。
func (s *Store) GetByIDs(ctx context.Context, ids []string) ([]domain.Article, error) {
	ids = domain.UniqueIDs(ids)
	if len(ids) == 0 {
		return nil, nil
	}

	payload, err := json.Marshal(map[string]any{
		"size":    len(ids),
		"_source": []string{"id", "tags", "kind", "level", "quality", "promo", "enriched_at"},
		"query":   map[string]any{"ids": map[string]any{"values": ids}},
	})
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
		return nil, fmt.Errorf("mget: %s", raw)
	}

	var parsed articleSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	articles := make([]domain.Article, 0, len(parsed.Hits.Hits))
	for _, hit := range parsed.Hits.Hits {
		article := hit.Source
		if article.ID == "" {
			article.ID = hit.ID
		}
		articles = append(articles, article)
	}
	return articles, nil
}

type articleHit struct {
	ID        string              `json:"_id"`
	Source    domain.Article      `json:"_source"`
	Highlight map[string][]string `json:"highlight"`
	Sort      []any               `json:"sort"`
}

type articleSearchResponse struct {
	Hits struct {
		Hits []articleHit `json:"hits"`
	} `json:"hits"`
}

// Search は記事を1ページ分取り、続きがあれば next カーソルを付ける。
func (s *Store) Search(ctx context.Context, query domain.FeedQuery) (domain.FeedPage, error) {
	payload, err := json.Marshal(articleSearchBody(query))
	if err != nil {
		return domain.FeedPage{}, err
	}

	res, err := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(articlesAlias),
		s.client.Search.WithBody(bytes.NewReader(payload)),
	)
	if err != nil {
		return domain.FeedPage{}, err
	}
	defer res.Body.Close()
	if res.IsError() {
		raw, _ := io.ReadAll(res.Body)
		return domain.FeedPage{}, fmt.Errorf("search: %s", raw)
	}

	var parsed articleSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return domain.FeedPage{}, err
	}

	return feedPageFromHits(parsed.Hits.Hits)
}

// feedPageFromHits は size+1 件のヒットをページサイズに切り、末尾の sort を next にする。
func feedPageFromHits(hits []articleHit) (domain.FeedPage, error) {
	var next string
	if len(hits) > domain.FeedPageSize {
		cursor, err := domain.EncodeSearchAfter(hits[domain.FeedPageSize-1].Sort)
		if err != nil {
			return domain.FeedPage{}, err
		}
		next = cursor
		hits = hits[:domain.FeedPageSize]
	}
	articles := make([]domain.Article, 0, len(hits))
	for _, hit := range hits {
		articles = append(articles, applyHighlight(hit.Source, hit.Highlight))
	}
	return domain.FeedPage{Articles: articles, Next: next}, nil
}

// applyHighlight は ES の highlight を title / summary の表示用フィールドへ写す。
func applyHighlight(article domain.Article, highlight map[string][]string) domain.Article {
	if titles := highlight["title"]; len(titles) > 0 {
		article.TitleHighlighted = titles[0]
	}
	if summaries := highlight["summary"]; len(summaries) > 0 {
		article.SummaryHighlighted = summaries[0]
	}
	return article
}

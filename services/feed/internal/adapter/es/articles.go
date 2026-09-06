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
	articlesVersion = 2
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
      "id": { "type": "keyword" },
      "source": { "type": "keyword" },
      "url": { "type": "keyword" },
      "title": { "type": "text", "analyzer": "ja_analyzer" },
      "summary": { "type": "text", "analyzer": "ja_analyzer" },
      "tags": { "type": "keyword" },
      "published_at": { "type": "date" }
    }
  }
}`

// articlesPhysicalIndex は alias が指す実インデックス名（articles-vN）。
func articlesPhysicalIndex() string {
	return fmt.Sprintf("%s-v%d", articlesAlias, articlesVersion)
}

// EnsureIndex は現行マッピングのインデックスを用意し、alias を張り替える。
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
		"_source": []string{"id", "tags"},
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

// articleSearchBody は一覧 / 検索の ES クエリを組む。size は1件多めにして次ページ判定する。
func articleSearchBody(query domain.FeedQuery) map[string]any {
	boolQuery := map[string]any{}
	if strings.TrimSpace(query.Text) == "" {
		boolQuery["must"] = []any{map[string]any{"match_all": map[string]any{}}}
	} else {
		boolQuery["must"] = []any{
			map[string]any{
				"multi_match": map[string]any{
					"query":    query.Text,
					"fields":   []string{"title^2", "summary"},
					"analyzer": "ja_analyzer",
				},
			},
		}
	}
	if len(query.FilterTags) > 0 {
		boolQuery["filter"] = []any{
			map[string]any{"terms": map[string]any{"tags": query.FilterTags}},
		}
	}
	if mustNot := articleMustNot(query); len(mustNot) > 0 {
		boolQuery["must_not"] = mustNot
	}
	if should := articleShould(query); len(should) > 0 {
		boolQuery["should"] = should
	}

	body := map[string]any{
		"size":  domain.FeedPageSize + 1,
		"sort":  articleSort(query.Sort),
		"query": map[string]any{"bool": boolQuery},
	}
	if len(query.SearchAfter) > 0 {
		body["search_after"] = query.SearchAfter
	}
	if strings.TrimSpace(query.Text) != "" {
		body["highlight"] = map[string]any{
			"encoder":   "html",
			"pre_tags":  []string{"<mark>"},
			"post_tags": []string{"</mark>"},
			"fields": map[string]any{
				"title":   map[string]any{"number_of_fragments": 0},
				"summary": map[string]any{"number_of_fragments": 0},
			},
		}
	}
	return body
}

// articleMustNot は除外タグと、履歴・ブックマーク済み ID を落とす条件。
func articleMustNot(query domain.FeedQuery) []any {
	var mustNot []any
	if len(query.ExcludeTags) > 0 {
		mustNot = append(mustNot, map[string]any{"terms": map[string]any{"tags": query.ExcludeTags}})
	}
	if len(query.ExcludeIDs) > 0 {
		mustNot = append(mustNot, map[string]any{"ids": map[string]any{"values": query.ExcludeIDs}})
	}
	return mustNot
}

// articleShould は関心・ブックマーク・履歴タグを boost 違いで載せる。
func articleShould(query domain.FeedQuery) []any {
	var should []any
	if len(query.InterestTags) > 0 {
		should = append(should, map[string]any{"terms": map[string]any{"tags": query.InterestTags, "boost": 4}})
	}
	if len(query.BookmarkTags) > 0 {
		should = append(should, map[string]any{"terms": map[string]any{"tags": query.BookmarkTags, "boost": 3}})
	}
	if len(query.HistoryTags) > 0 {
		should = append(should, map[string]any{"terms": map[string]any{"tags": query.HistoryTags, "boost": 2}})
	}
	return should
}

// articleSort はおすすめならスコア優先、それ以外は新しい順。id は search_after のタイブレーク。
func articleSort(sort string) []any {
	tie := map[string]any{"id": map[string]any{"order": "desc"}}
	published := map[string]any{"published_at": map[string]any{"order": "desc"}}
	if sort == domain.SortRecommended {
		return []any{
			map[string]any{"_score": map[string]any{"order": "desc"}},
			published,
			tie,
		}
	}
	return []any{published, tie}
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

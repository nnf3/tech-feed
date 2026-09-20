package es

import (
	"strings"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

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

// articleMustNot は一覧から落とす条件。
//   - 除外タグに一致する記事
//   - 履歴・ブックマーク済みの記事 ID
//   - おすすめソート時のみ、promo が true の宣伝記事
func articleMustNot(query domain.FeedQuery) []any {
	var mustNot []any
	if len(query.ExcludeTags) > 0 {
		mustNot = append(mustNot, map[string]any{"terms": map[string]any{"tags": query.ExcludeTags}})
	}
	if len(query.ExcludeIDs) > 0 {
		mustNot = append(mustNot, map[string]any{"ids": map[string]any{"values": query.ExcludeIDs}})
	}
	if query.Sort == domain.SortRecommended {
		mustNot = append(mustNot, map[string]any{"term": map[string]any{"promo": true}})
	}
	return mustNot
}

// articleShould はおすすめ用の加点条件。
//   - 関心タグ: boost 4
//   - ブックマーク由来タグ: boost 3
//   - 履歴由来タグ: boost 2
//   - おすすめソート時のみ、quality >= 1 を boost 2
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
	if query.Sort == domain.SortRecommended {
		should = append(should, map[string]any{"range": map[string]any{"quality": map[string]any{"gte": 1.0, "boost": 2}}})
	}
	return should
}

// articleSort はページ内の並び。id は search_after のタイブレーク。
//   - おすすめ: _score desc → published_at desc → id desc
//   - それ以外: published_at desc → id desc
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

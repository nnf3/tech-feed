package domain

import (
	"sort"
	"strings"
)

// UniqueIDs は複数スライスの ID を順に集め、空と重複を除く。
func UniqueIDs(ids ...[]string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, group := range ids {
		for _, id := range group {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}
	return out
}

// ArticlesByID は ID から記事を引くためのマップを作る。
func ArticlesByID(articles []Article) map[string]Article {
	out := make(map[string]Article, len(articles))
	for _, article := range articles {
		if article.ID == "" {
			continue
		}
		out[article.ID] = article
	}
	return out
}

// BookmarkTagWeights は保存記事のタグを1回ずつ並べる。出現回数がいちばんの重みになる。
func BookmarkTagWeights(byID map[string]Article, items []Bookmark) []string {
	var out []string
	for _, item := range items {
		article, ok := byID[item.ArticleID]
		if !ok {
			continue
		}
		out = append(out, article.Tags...)
	}
	return out
}

// HistoryTagWeights は閲覧回数ぶんタグを重ねる。よく見た話題を強くする。
func HistoryTagWeights(byID map[string]Article, items []HistoryEntry) []string {
	var out []string
	for _, item := range items {
		article, ok := byID[item.ArticleID]
		if !ok {
			continue
		}
		n := item.ViewCount
		if n < 1 {
			n = 1
		}
		if n > RecommendHistoryViewCap {
			n = RecommendHistoryViewCap
		}
		for i := 0; i < n; i++ {
			out = append(out, article.Tags...)
		}
	}
	return out
}

// TopTags は重み付きタグから除外分を除き、頻度の高い順に limit 件返す。
func TopTags(weighted []string, limit int, exclude []string) []string {
	skip := make(map[string]struct{}, len(exclude))
	for _, tag := range exclude {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag != "" {
			skip[tag] = struct{}{}
		}
	}
	counts := map[string]int{}
	for _, raw := range weighted {
		tag := strings.ToLower(strings.TrimSpace(raw))
		if tag == "" {
			continue
		}
		if _, ok := skip[tag]; ok {
			continue
		}
		counts[tag]++
	}
	type pair struct {
		tag string
		n   int
	}
	ranked := make([]pair, 0, len(counts))
	for tag, n := range counts {
		ranked = append(ranked, pair{tag: tag, n: n})
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].n != ranked[j].n {
			return ranked[i].n > ranked[j].n
		}
		return ranked[i].tag < ranked[j].tag
	})
	if limit <= 0 || limit > len(ranked) {
		limit = len(ranked)
	}
	out := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, ranked[i].tag)
	}
	return out
}

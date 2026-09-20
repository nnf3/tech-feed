package es

import (
	"testing"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

func TestArticleSearchBodyFilterTags(t *testing.T) {
	body := articleSearchBody(domain.FeedQuery{FilterTags: []string{"react"}})
	query, _ := body["query"].(map[string]any)
	boolQuery, _ := query["bool"].(map[string]any)
	filters, ok := boolQuery["filter"].([]any)
	if !ok || len(filters) == 0 {
		t.Fatalf("filter missing: %#v", boolQuery)
	}
	terms, _ := filters[0].(map[string]any)["terms"].(map[string]any)
	tags, _ := terms["tags"].([]string)
	if len(tags) != 1 || tags[0] != "react" {
		t.Fatalf("tags: %#v", terms["tags"])
	}
}

func TestArticleSearchBodyHighlight(t *testing.T) {
	if _, ok := articleSearchBody(domain.FeedQuery{})["highlight"]; ok {
		t.Fatal("highlight should be off without text")
	}
	body := articleSearchBody(domain.FeedQuery{Text: "go"})
	highlight, ok := body["highlight"].(map[string]any)
	if !ok {
		t.Fatal("highlight missing")
	}
	if highlight["encoder"] != "html" {
		t.Fatalf("encoder: %#v", highlight["encoder"])
	}
}

func TestArticleSearchBodyRecommendSignals(t *testing.T) {
	body := articleSearchBody(domain.FeedQuery{
		InterestTags: []string{"go"},
		BookmarkTags: []string{"k8s"},
		HistoryTags:  []string{"rust"},
		ExcludeTags:  []string{"poem"},
		ExcludeIDs:   []string{"seen"},
	})
	query, _ := body["query"].(map[string]any)
	boolQuery, _ := query["bool"].(map[string]any)
	should, _ := boolQuery["should"].([]any)
	if len(should) != 3 {
		t.Fatalf("should: %#v", should)
	}
	mustNot, _ := boolQuery["must_not"].([]any)
	if len(mustNot) != 2 {
		t.Fatalf("must_not: %#v", mustNot)
	}
}

func TestArticleSearchBodyRecommendedSignals(t *testing.T) {
	body := articleSearchBody(domain.FeedQuery{
		Sort:         domain.SortRecommended,
		InterestTags: []string{"go"},
	})
	query, _ := body["query"].(map[string]any)
	boolQuery, _ := query["bool"].(map[string]any)
	should, _ := boolQuery["should"].([]any)
	if len(should) != 2 {
		t.Fatalf("should: %#v", should)
	}
	mustNot, _ := boolQuery["must_not"].([]any)
	if len(mustNot) != 1 {
		t.Fatalf("must_not: %#v", mustNot)
	}
	term, _ := mustNot[0].(map[string]any)["term"].(map[string]any)
	if term["promo"] != true {
		t.Fatalf("promo filter: %#v", mustNot)
	}
}

func TestArticleSearchBodySortAndCursor(t *testing.T) {
	newest := articleSearchBody(domain.FeedQuery{})
	if newest["size"] != domain.FeedPageSize+1 {
		t.Fatalf("size: %#v", newest["size"])
	}
	if _, ok := newest["search_after"]; ok {
		t.Fatal("search_after should be off on first page")
	}
	sort, _ := newest["sort"].([]any)
	if len(sort) != 2 {
		t.Fatalf("new sort: %#v", newest["sort"])
	}
	if _, ok := sort[0].(map[string]any)["published_at"]; !ok {
		t.Fatalf("new sort should start with published_at: %#v", sort)
	}

	rec := articleSearchBody(domain.FeedQuery{Sort: domain.SortRecommended, SearchAfter: []any{1, "x"}})
	recSort, _ := rec["sort"].([]any)
	if _, ok := recSort[0].(map[string]any)["_score"]; !ok {
		t.Fatalf("recommended sort: %#v", recSort)
	}
	after, _ := rec["search_after"].([]any)
	if len(after) != 2 {
		t.Fatalf("search_after: %#v", rec["search_after"])
	}
}

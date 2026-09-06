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

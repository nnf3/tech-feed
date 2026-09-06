package zenn

import (
	"context"
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

var htmlTag = regexp.MustCompile(`<[^>]+>`)

type Source struct {
	URL    string
	parser *gofeed.Parser
}

func New(feedURL string) *Source {
	if feedURL == "" {
		feedURL = "https://zenn.dev/feed"
	}
	return &Source{URL: feedURL, parser: gofeed.NewParser()}
}

func (z *Source) Fetch(ctx context.Context) ([]domain.Article, error) {
	feed, err := z.parser.ParseURLWithContext(z.URL, ctx)
	if err != nil {
		return nil, err
	}

	out := make([]domain.Article, 0, len(feed.Items))
	for _, item := range feed.Items {
		if item.Link == "" || item.Title == "" {
			continue
		}
		published := time.Now().UTC()
		if item.PublishedParsed != nil {
			published = item.PublishedParsed.UTC()
		} else if item.UpdatedParsed != nil {
			published = item.UpdatedParsed.UTC()
		}

		tags := normalizeTags(item.Categories)

		out = append(out, domain.Article{
			ID:          domain.IDFromURL(item.Link),
			Source:      "zenn",
			URL:         item.Link,
			Title:       strings.TrimSpace(item.Title),
			Summary:     summarize(item.Description),
			Tags:        tags,
			PublishedAt: published,
		})
	}
	return out, nil
}

func normalizeTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, raw := range tags {
		tag := strings.ToLower(strings.TrimSpace(raw))
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	return out
}

func summarize(raw string) string {
	text := htmlTag.ReplaceAllString(raw, " ")
	text = html.UnescapeString(text)
	text = strings.Join(strings.Fields(text), " ")
	if len([]rune(text)) > 180 {
		return string([]rune(text)[:180]) + "…"
	}
	return text
}

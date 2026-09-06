package crawler

import (
	"crypto/sha256"
	"encoding/hex"
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/nnf3/tech-feed/services/feed/internal/article"
)

var htmlTag = regexp.MustCompile(`<[^>]+>`)

type Zenn struct {
	URL    string
	parser *gofeed.Parser
}

func NewZenn(feedURL string) *Zenn {
	if feedURL == "" {
		feedURL = "https://zenn.dev/feed"
	}
	return &Zenn{URL: feedURL, parser: gofeed.NewParser()}
}

func (z *Zenn) Fetch() ([]article.Article, error) {
	feed, err := z.parser.ParseURL(z.URL)
	if err != nil {
		return nil, err
	}

	out := make([]article.Article, 0, len(feed.Items))
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

		tags := item.Categories
		if tags == nil {
			tags = []string{}
		}

		out = append(out, article.Article{
			ID:          idFromURL(item.Link),
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

func idFromURL(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:16])
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

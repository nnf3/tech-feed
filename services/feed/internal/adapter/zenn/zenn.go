package zenn

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

const (
	defaultFeedURL = "https://zenn.dev/feed"
	defaultAPIBase = "https://zenn.dev"
	topicWorkers   = 5
)

var htmlTag = regexp.MustCompile(`<[^>]+>`)

type Source struct {
	URL    string
	API    string
	client *http.Client
	parser *gofeed.Parser
}

// New は Zenn ソースを作る。feedURL が空なら公式 RSS を使う。
func New(feedURL string) *Source {
	if feedURL == "" {
		feedURL = defaultFeedURL
	}
	return &Source{
		URL:    feedURL,
		API:    defaultAPIBase,
		client: &http.Client{Timeout: 10 * time.Second},
		parser: gofeed.NewParser(),
	}
}

// Fetch は RSS から記事一覧を取り、各記事のトピックを API で補う。
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

		out = append(out, domain.Article{
			ID:          domain.IDFromURL(item.Link),
			Source:      "zenn",
			URL:         item.Link,
			Title:       strings.TrimSpace(item.Title),
			Summary:     summarize(item.Description),
			Tags:        normalizeTags(item.Categories),
			PublishedAt: published,
		})
	}

	z.enrichTopics(ctx, out)
	return out, nil
}

// enrichTopics は RSS に無いトピックを、記事/本の詳細 API から並行取得して tags に入れる。
func (z *Source) enrichTopics(ctx context.Context, articles []domain.Article) {
	sem := make(chan struct{}, topicWorkers)
	var wg sync.WaitGroup
	for i := range articles {
		kind, slug, ok := resourceFromURL(articles[i].URL)
		if !ok {
			continue
		}
		wg.Add(1)
		go func(i int, kind, slug string) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				return
			}
			tags, err := z.fetchTopics(ctx, kind, slug)
			if err != nil {
				log.Printf("zenn topics %s/%s: %v", kind, slug, err)
				return
			}
			if len(tags) == 0 {
				return
			}
			articles[i].Tags = tags
		}(i, kind, slug)
	}
	wg.Wait()
}

// fetchTopics は /api/articles/{slug} または /api/books/{slug} から topics を取る。
func (z *Source) fetchTopics(ctx context.Context, kind, slug string) ([]string, error) {
	endpoint := strings.TrimRight(z.API, "/") + "/api/" + kind + "/" + url.PathEscape(slug)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "tech-feed/0.1 (https://github.com/nnf3/tech-feed)")

	res, err := z.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("article api %s", res.Status)
	}
	return topicsFromJSON(body)
}

type topicName struct {
	Name string `json:"name"`
}

type resourceDetail struct {
	Article struct {
		Topics []topicName `json:"topics"`
	} `json:"article"`
	Book struct {
		Topics []topicName `json:"topics"`
	} `json:"book"`
}

// topicsFromJSON は詳細 API の JSON から topic 名だけ抜き出す。
func topicsFromJSON(raw []byte) ([]string, error) {
	var detail resourceDetail
	if err := json.Unmarshal(raw, &detail); err != nil {
		return nil, err
	}
	topics := detail.Article.Topics
	if len(topics) == 0 {
		topics = detail.Book.Topics
	}
	names := make([]string, 0, len(topics))
	for _, topic := range topics {
		names = append(names, topic.Name)
	}
	return normalizeTags(names), nil
}

// resourceFromURL は記事 URL から articles/books と slug を取り出す。
func resourceFromURL(raw string) (kind, slug string, ok bool) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", "", false
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	for i := 0; i < len(parts)-1; i++ {
		if (parts[i] == "articles" || parts[i] == "books") && parts[i+1] != "" {
			return parts[i], parts[i+1], true
		}
	}
	return "", "", false
}

// normalizeTags は前後空白を落とし、小文字化して重複を除く。
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

// summarize は RSS の description から HTML を除き、180 文字に切る。
func summarize(raw string) string {
	text := htmlTag.ReplaceAllString(raw, " ")
	text = html.UnescapeString(text)
	text = strings.Join(strings.Fields(text), " ")
	if len([]rune(text)) > 180 {
		return string([]rune(text)[:180]) + "…"
	}
	return text
}

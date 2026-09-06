package qiita

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nnf3/tech-feed/services/feed/internal/adapter/feedutil"
	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

const defaultAPI = "https://qiita.com/api/v2/items?per_page=20"

type Source struct {
	URL    string
	client *http.Client
}

// New は Qiita ソースを作る。apiURL が空なら公式の最新記事 API を使う。
func New(apiURL string) *Source {
	if apiURL == "" {
		apiURL = defaultAPI
	}
	return &Source{
		URL:    apiURL,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type item struct {
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	Tags      []struct {
		Name string `json:"name"`
	} `json:"tags"`
}

// Fetch は Qiita API から最新記事とタグを取る。RSS には category が無い。
func (q *Source) Fetch(ctx context.Context) ([]domain.Article, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, q.URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "tech-feed/0.1 (https://github.com/nnf3/tech-feed)")

	res, err := q.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("qiita api %s", res.Status)
	}

	var items []item
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, err
	}

	out := make([]domain.Article, 0, len(items))
	for _, it := range items {
		if it.URL == "" || it.Title == "" {
			continue
		}
		names := make([]string, 0, len(it.Tags))
		for _, tag := range it.Tags {
			names = append(names, tag.Name)
		}
		published := it.CreatedAt.UTC()
		if published.IsZero() {
			published = time.Now().UTC()
		}
		out = append(out, domain.Article{
			ID:          domain.IDFromURL(it.URL),
			Source:      "qiita",
			URL:         it.URL,
			Title:       strings.TrimSpace(it.Title),
			Summary:     feedutil.Summarize(it.Body),
			Tags:        feedutil.NormalizeTags(names),
			PublishedAt: published,
		})
	}
	return out, nil
}

package typesafe

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

const (
	defaultBaseURL        = "https://api.typesafe.ai"
	defaultModel          = "jev-latest"
	defaultConcurrency    = 4    // 記事ごとの並列リクエスト数
	defaultMinConfidence  = 0.45 // Choice / Score を採用する下限。未満なら kind / level / quality は書かない
	defaultPromoThreshold = 0.7  // promo の Noul がこの値以上なら宣伝記事とみなす
)

type Config struct {
	APIKey         string
	BaseURL        string
	Model          string
	Concurrency    int
	MinConfidence  float64
	PromoThreshold float64
	HTTPClient     *http.Client
}

type Client struct {
	apiKey         string
	endpoint       string
	model          string
	concurrency    int
	minConfidence  float64
	promoThreshold float64
	http           *http.Client
}

func NewFromEnv() domain.ArticleEnricher {
	key := strings.TrimSpace(os.Getenv("TYPESAFE_API_KEY"))
	if key == "" {
		log.Printf("typesafe: TYPESAFE_API_KEY unset, skip article enrichment")
		return domain.NoopArticleEnricher()
	}
	return New(Config{
		APIKey:  key,
		BaseURL: os.Getenv("TYPESAFE_BASE_URL"),
		Model:   os.Getenv("TYPESAFE_MODEL"),
	})
}

func New(cfg Config) *Client {
	base := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if base == "" {
		base = defaultBaseURL
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = defaultModel
	}
	concurrency := cfg.Concurrency
	if concurrency <= 0 {
		concurrency = defaultConcurrency
	}
	minConfidence := cfg.MinConfidence
	if minConfidence <= 0 {
		minConfidence = defaultMinConfidence
	}
	promoThreshold := cfg.PromoThreshold
	if promoThreshold <= 0 {
		promoThreshold = defaultPromoThreshold
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	return &Client{
		apiKey:         cfg.APIKey,
		endpoint:       base + "/v1/systemone",
		model:          model,
		concurrency:    concurrency,
		minConfidence:  minConfidence,
		promoThreshold: promoThreshold,
		http:           httpClient,
	}
}

// Enrich は記事ごとに TypeSafe を並列で呼び、判定できたものだけ kind / level などを書く。
// 失敗した記事は元のまま残し、最初のエラーを返す。
func (c *Client) Enrich(ctx context.Context, articles []domain.Article) ([]domain.Article, error) {
	if len(articles) == 0 {
		return articles, nil
	}
	out := append([]domain.Article(nil), articles...)
	sem := make(chan struct{}, c.concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for i := range out {
		article := out[i]
		sem <- struct{}{}
		wg.Add(1)
		go func(i int, article domain.Article) {
			defer wg.Done()
			defer func() { <-sem }()
			got, err := c.evaluate(ctx, article)
			if err != nil {
				log.Printf("typesafe enrich skipped %s: %v", article.ID, err)
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}
			mu.Lock()
			out[i] = got
			mu.Unlock()
		}(i, article)
	}
	wg.Wait()
	if err := ctx.Err(); err != nil {
		return out, err
	}
	return out, firstErr
}

// evaluate は1記事を state にして質問を投げ、answer を Article のフィールドへ写す。
func (c *Client) evaluate(ctx context.Context, article domain.Article) (domain.Article, error) {
	payload, err := json.Marshal(systemOneRequest{
		State: map[string]any{
			"title":   article.Title,
			"summary": article.Summary,
			"tags":    article.Tags,
			"source":  article.Source,
		},
		Model:     c.model,
		Questions: enrichQuestions,
	})
	if err != nil {
		return article, err
	}

	body, err := c.post(ctx, payload)
	if err != nil {
		return article, err
	}

	var resp systemOneResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return article, fmt.Errorf("decode typesafe response: %w", err)
	}
	applyAnswers(&article, resp, c.minConfidence, c.promoThreshold)
	return article, nil
}

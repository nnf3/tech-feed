package usecase

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"
)

const defaultCrawlInterval = 15 * time.Minute

// Crawler は起動時と一定間隔で ingest を回す。cmd/crawler から別プロセスで動かす。
type Crawler struct {
	ingest  *Ingest
	every   time.Duration
	timeout time.Duration
	mu      sync.Mutex
}

func NewCrawler(ingest *Ingest, every, timeout time.Duration) *Crawler {
	if timeout <= 0 {
		timeout = time.Minute
	}
	return &Crawler{ingest: ingest, every: every, timeout: timeout}
}

// ParseInterval は CRAWL_INTERVAL を読む。空は 15m。0 / off は定期実行なし。
func ParseInterval(raw string) time.Duration {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return defaultCrawlInterval
	}
	if raw == "0" || raw == "off" || raw == "false" {
		return 0
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d < 0 {
		log.Printf("invalid CRAWL_INTERVAL %q, using %s", raw, defaultCrawlInterval)
		return defaultCrawlInterval
	}
	return d
}

// Start は先に 1 回取り込み、間隔が正なら ticker で繰り返す。ctx が切れたら止まる。
func (c *Crawler) Start(ctx context.Context) {
	c.Run(ctx, "startup")
	if c.every <= 0 {
		log.Printf("periodic crawl disabled")
		return
	}
	log.Printf("periodic crawl every %s", c.every)
	ticker := time.NewTicker(c.every)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.Run(ctx, "periodic")
		}
	}
}

// Run は進行中の取り込みがあればスキップする。
func (c *Crawler) Run(parent context.Context, reason string) {
	if !c.mu.TryLock() {
		log.Printf("crawl skipped (%s): already running", reason)
		return
	}
	defer c.mu.Unlock()

	ctx, cancel := context.WithTimeout(parent, c.timeout)
	defer cancel()
	items, err := c.ingest.Run(ctx)
	if err != nil {
		log.Printf("crawl %s skipped: %v", reason, err)
		return
	}
	log.Printf("crawl %s ingested %d articles", reason, len(items))
}

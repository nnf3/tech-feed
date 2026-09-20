package typesafe

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// maxAttempts は 429 / 529 / 5xx を含む最大試行回数。
const maxAttempts = 3

type systemOneRequest struct {
	State     map[string]any      `json:"state"`
	Model     string              `json:"model"`
	Questions map[string]question `json:"questions"`
}

func (c *Client) post(ctx context.Context, payload []byte) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			delay := time.Duration(200*(1<<attempt)) * time.Millisecond
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			case <-timer.C:
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "tech-feed/0.1 (https://github.com/nnf3/tech-feed)")

		res, err := c.http.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		raw, readErr := io.ReadAll(io.LimitReader(res.Body, 2<<20))
		res.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if res.StatusCode == http.StatusTooManyRequests || res.StatusCode == 529 || res.StatusCode >= 500 {
			lastErr = fmt.Errorf("typesafe %s: %s", res.Status, truncate(raw))
			continue
		}
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			return nil, fmt.Errorf("typesafe %s: %s", res.Status, truncate(raw))
		}
		return raw, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("typesafe request failed")
	}
	return nil, lastErr
}

func truncate(raw []byte) string {
	const limit = 240
	s := strings.TrimSpace(string(raw))
	if len(s) <= limit {
		return s
	}
	return s[:limit] + "…"
}

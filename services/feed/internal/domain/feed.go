package domain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	FeedPageSize    = 20
	SortNew         = "new"
	SortRecommended = "recommended"
)

type FeedPage struct {
	Articles []Article `json:"articles"`
	Next     string    `json:"next,omitempty"`
}

func NormalizeFeedSort(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), SortRecommended) {
		return SortRecommended
	}
	return SortNew
}

func EncodeSearchAfter(values []any) (string, error) {
	if len(values) == 0 {
		return "", nil
	}
	raw, err := json.Marshal(values)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func DecodeSearchAfter(raw string) ([]any, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("%w", ErrInvalidSearchAfter)
	}
	var values []any
	if err := json.Unmarshal(b, &values); err != nil || len(values) == 0 {
		return nil, fmt.Errorf("%w", ErrInvalidSearchAfter)
	}
	return values, nil
}

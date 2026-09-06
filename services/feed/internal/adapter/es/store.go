package es

import (
	"context"
	"fmt"
	"time"

	elasticsearch "github.com/elastic/go-elasticsearch/v9"
)

type Store struct {
	client *elasticsearch.Client
}

func New(esHost string) (*Store, error) {
	if esHost == "" {
		esHost = "http://localhost:9200"
	}
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{esHost},
	})
	if err != nil {
		return nil, fmt.Errorf("create elasticsearch client: %w", err)
	}
	return &Store{client: client}, nil
}

func (s *Store) WaitReady(ctx context.Context) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		res, err := s.client.Info(s.client.Info.WithContext(ctx))
		if err == nil {
			res.Body.Close()
			if res.StatusCode < 300 {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			if err != nil {
				return fmt.Errorf("elasticsearch not ready: %w", err)
			}
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

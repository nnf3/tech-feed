package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (s *Store) existsAlias(ctx context.Context, name string) (bool, error) {
	res, err := s.client.Indices.ExistsAlias([]string{name}, s.client.Indices.ExistsAlias.WithContext(ctx))
	if err != nil {
		return false, err
	}
	res.Body.Close()
	return res.StatusCode == http.StatusOK, nil
}

func (s *Store) existsIndex(ctx context.Context, name string) (bool, error) {
	res, err := s.client.Indices.Exists([]string{name}, s.client.Indices.Exists.WithContext(ctx))
	if err != nil {
		return false, err
	}
	res.Body.Close()
	return res.StatusCode == http.StatusOK, nil
}

func (s *Store) aliasTargets(ctx context.Context, alias string) ([]string, error) {
	res, err := s.client.Indices.GetAlias(
		s.client.Indices.GetAlias.WithContext(ctx),
		s.client.Indices.GetAlias.WithName(alias),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("get alias %s: %s", alias, body)
	}

	var payload map[string]json.RawMessage
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, err
	}

	targets := make([]string, 0, len(payload))
	for index := range payload {
		targets = append(targets, index)
	}
	return targets, nil
}

func (s *Store) createIndex(ctx context.Context, name, mapping string) error {
	res, err := s.client.Indices.Create(
		name,
		s.client.Indices.Create.WithContext(ctx),
		s.client.Indices.Create.WithBody(strings.NewReader(mapping)),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("create index %s: %s", name, body)
	}
	return nil
}

func (s *Store) deleteIndex(ctx context.Context, name string) error {
	res, err := s.client.Indices.Delete([]string{name}, s.client.Indices.Delete.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil
	}
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("delete index %s: %s", name, body)
	}
	return nil
}

func (s *Store) reindex(ctx context.Context, source, dest string) error {
	body, err := json.Marshal(map[string]any{
		"source": map[string]any{"index": source},
		"dest":   map[string]any{"index": dest},
	})
	if err != nil {
		return err
	}

	res, err := s.client.Reindex(
		bytes.NewReader(body),
		s.client.Reindex.WithContext(ctx),
		s.client.Reindex.WithWaitForCompletion(true),
		s.client.Reindex.WithRefresh(true),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		raw, _ := io.ReadAll(res.Body)
		return fmt.Errorf("reindex %s -> %s: %s", source, dest, raw)
	}
	return nil
}

func (s *Store) swapAlias(ctx context.Context, alias, desired string, current []string) error {
	actions := make([]map[string]any, 0, len(current)+1)
	already := false
	for _, old := range current {
		if old == desired {
			already = true
			continue
		}
		actions = append(actions, map[string]any{
			"remove": map[string]any{"index": old, "alias": alias},
		})
	}
	if !already {
		actions = append(actions, map[string]any{
			"add": map[string]any{"index": desired, "alias": alias},
		})
	}
	if len(actions) == 0 {
		return nil
	}

	body, err := json.Marshal(map[string]any{"actions": actions})
	if err != nil {
		return err
	}

	res, err := s.client.Indices.UpdateAliases(bytes.NewReader(body), s.client.Indices.UpdateAliases.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		raw, _ := io.ReadAll(res.Body)
		return fmt.Errorf("swap alias %s -> %s: %s", alias, desired, raw)
	}
	return nil
}

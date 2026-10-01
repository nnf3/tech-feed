package es

import (
	"context"
	"fmt"
	"log"
)

const (
	articlesAlias   = "articles"
	articlesVersion = 4
)

// articlesMapping は実インデックスの定義。
// 変えるときは articlesVersion を上げる。起動時に新インデックスへ reindex して alias を張り替える。
const articlesMapping = `{
  "settings": {
    "analysis": {
      "analyzer": {
        "ja_analyzer": {
          "type": "custom",
          "tokenizer": "kuromoji_tokenizer",
          "filter": ["kuromoji_baseform", "lowercase", "icu_normalizer"]
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "id": { "type": "keyword" },
      "source": { "type": "keyword" },
      "url": { "type": "keyword" },
      "title": { "type": "text", "analyzer": "ja_analyzer" },
      "summary": { "type": "text", "analyzer": "ja_analyzer" },
      "tags": { "type": "keyword" },
      "kind": { "type": "keyword" },
      "level": { "type": "keyword" },
      "quality": { "type": "float" },
      "promo": { "type": "boolean" },
      "enriched_at": { "type": "date" },
      "published_at": { "type": "date" }
    }
  }
}`

// articlesPhysicalIndex は alias が指す実インデックス名（articles-vN）。
func articlesPhysicalIndex() string {
	return fmt.Sprintf("%s-v%d", articlesAlias, articlesVersion)
}

// EnsureIndex は現行マッピングのインデックスを用意し、alias を張り替える。
func (s *Store) EnsureIndex(ctx context.Context) error {
	desired := articlesPhysicalIndex()

	exists, err := s.existsIndex(ctx, desired)
	if err != nil {
		return err
	}
	if !exists {
		log.Printf("elasticsearch: create index %s", desired)
		if err := s.createIndex(ctx, desired, articlesMapping); err != nil {
			return err
		}
	}

	isAlias, err := s.existsAlias(ctx, articlesAlias)
	if err != nil {
		return err
	}
	if !isAlias {
		// 旧実装は alias と同名の実インデックスを作っていた。名前を空けるため、ここだけ一瞬 404 になり得る。
		legacy, err := s.existsIndex(ctx, articlesAlias)
		if err != nil {
			return err
		}
		if legacy {
			log.Printf("elasticsearch: migrate concrete index %s -> %s", articlesAlias, desired)
			if err := s.reindex(ctx, articlesAlias, desired); err != nil {
				return err
			}
			if err := s.deleteIndex(ctx, articlesAlias); err != nil {
				return err
			}
		}
		log.Printf("elasticsearch: point alias %s -> %s", articlesAlias, desired)
		return s.swapAlias(ctx, articlesAlias, desired, nil)
	}

	current, err := s.aliasTargets(ctx, articlesAlias)
	if err != nil {
		return err
	}
	if len(current) == 1 && current[0] == desired {
		return nil
	}

	log.Printf("elasticsearch: swap alias %s from %v to %s", articlesAlias, current, desired)
	for _, old := range current {
		if old == desired {
			continue
		}
		if err := s.reindex(ctx, old, desired); err != nil {
			return err
		}
	}
	if err := s.swapAlias(ctx, articlesAlias, desired, current); err != nil {
		return err
	}
	for _, old := range current {
		if old == desired {
			continue
		}
		if err := s.deleteIndex(ctx, old); err != nil {
			return err
		}
	}
	return nil
}

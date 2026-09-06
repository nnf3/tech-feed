package domain

import "context"

// ArticleIndex は永続化のポート。いまの実装は Elasticsearch。
type ArticleIndex interface {
	BulkUpsert(ctx context.Context, articles []Article) error
	Search(ctx context.Context, query string) ([]Article, error)
}

// Source は外部メディアから記事を取得する。
type Source interface {
	Fetch(ctx context.Context) ([]Article, error)
}

// Ranker はフィードの並び順を決める。
// いまは新しい順のローカル実装。後から recommendation サービスに差し替えられる。
type Ranker interface {
	Rank(articles []Article) []Article
}

// UserStore はユーザーの永続化。いまの実装は Postgres。
type UserStore interface {
	Upsert(ctx context.Context, user User) error
	Get(ctx context.Context, id string) (*User, error)
}

package domain

import "context"

// FeedQuery は一覧 / 検索の条件。プロフィールのタグは ES の must / should に載せる。
type FeedQuery struct {
	Text         string
	FilterTags   []string
	InterestTags []string
	ExcludeTags  []string
	Sort         string
	SearchAfter  []any
}

// ArticleIndex は永続化のポート。いまの実装は Elasticsearch。
type ArticleIndex interface {
	BulkUpsert(ctx context.Context, articles []Article) error
	Search(ctx context.Context, query FeedQuery) (FeedPage, error)
}

// Source は外部メディアから記事を取得する。
type Source interface {
	Fetch(ctx context.Context) ([]Article, error)
}

// UserStore はユーザーの永続化。いまの実装は Postgres。
type UserStore interface {
	Upsert(ctx context.Context, user User) error
	Get(ctx context.Context, id string) (*User, error)
}

// ProfileStore はプロフィールの永続化。いまの実装は Postgres。
type ProfileStore interface {
	UpsertProfile(ctx context.Context, profile Profile) error
	GetProfile(ctx context.Context, userID string) (*Profile, error)
}

// BookmarkStore はブックマークの永続化。いまの実装は Postgres。
type BookmarkStore interface {
	UpsertBookmark(ctx context.Context, bookmark Bookmark) error
	GetBookmark(ctx context.Context, userID, articleID string) (*Bookmark, error)
	DeleteBookmark(ctx context.Context, userID, articleID string) error
	ListBookmarks(ctx context.Context, userID string) ([]Bookmark, error)
	CountBookmarks(ctx context.Context, userID string) (int, error)
}

// HistoryStore は閲覧履歴の永続化。いまの実装は Postgres。
type HistoryStore interface {
	UpsertHistory(ctx context.Context, entry HistoryEntry) error
	GetHistory(ctx context.Context, userID, articleID string) (*HistoryEntry, error)
	ListHistory(ctx context.Context, userID string) ([]HistoryEntry, error)
	TrimHistory(ctx context.Context, userID string, keep int) error
}

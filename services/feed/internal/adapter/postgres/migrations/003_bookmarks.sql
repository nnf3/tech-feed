CREATE TABLE IF NOT EXISTS bookmarks (
    user_id TEXT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    article_id TEXT NOT NULL,
    url TEXT NOT NULL,
    title TEXT NOT NULL,
    source TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, article_id)
);

CREATE INDEX IF NOT EXISTS bookmarks_user_created_idx ON bookmarks (user_id, created_at DESC);

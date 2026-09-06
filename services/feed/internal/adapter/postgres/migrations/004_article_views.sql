CREATE TABLE IF NOT EXISTS article_views (
    user_id TEXT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    article_id TEXT NOT NULL,
    url TEXT NOT NULL,
    title TEXT NOT NULL,
    source TEXT NOT NULL DEFAULT '',
    viewed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    view_count INT NOT NULL DEFAULT 1,
    PRIMARY KEY (user_id, article_id)
);

CREATE INDEX IF NOT EXISTS article_views_user_viewed_idx ON article_views (user_id, viewed_at DESC);

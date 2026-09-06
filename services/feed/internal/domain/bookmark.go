package domain

import "time"

const MaxBookmarks = 200

type Bookmark struct {
	UserID    string    `json:"user_id"`
	ArticleID string    `json:"article_id"`
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
}

func (b Bookmark) Normalized() (Bookmark, error) {
	snap, err := normalizeArticleSnapshot(b.UserID, b.ArticleID, b.URL, b.Title, b.Source)
	if err != nil {
		return Bookmark{}, err
	}
	b.UserID = snap.userID
	b.ArticleID = snap.articleID
	b.URL = snap.url
	b.Title = snap.title
	b.Source = snap.source
	return b, nil
}

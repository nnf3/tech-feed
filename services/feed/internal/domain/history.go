package domain

import "time"

const MaxHistory = 500

type HistoryEntry struct {
	UserID    string    `json:"user_id"`
	ArticleID string    `json:"article_id"`
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	Source    string    `json:"source"`
	ViewedAt  time.Time `json:"viewed_at"`
	ViewCount int       `json:"view_count"`
}

func (h HistoryEntry) Normalized() (HistoryEntry, error) {
	snap, err := normalizeArticleSnapshot(h.UserID, h.ArticleID, h.URL, h.Title, h.Source)
	if err != nil {
		return HistoryEntry{}, err
	}
	h.UserID = snap.userID
	h.ArticleID = snap.articleID
	h.URL = snap.url
	h.Title = snap.title
	h.Source = snap.source
	if h.ViewCount < 1 {
		h.ViewCount = 1
	}
	return h, nil
}

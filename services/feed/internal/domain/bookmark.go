package domain

import (
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MaxBookmarks  = 200
	MaxTitleRunes = 300
	MaxSourceLen  = 32
)

type Bookmark struct {
	UserID    string    `json:"user_id"`
	ArticleID string    `json:"article_id"`
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
}

func (b Bookmark) Normalized() (Bookmark, error) {
	b.UserID = strings.TrimSpace(b.UserID)
	if b.UserID == "" {
		return Bookmark{}, ErrUserIDRequired
	}

	b.URL = strings.TrimSpace(b.URL)
	parsed, err := url.Parse(b.URL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return Bookmark{}, ErrInvalidArticleURL
	}

	b.ArticleID = strings.TrimSpace(b.ArticleID)
	if b.ArticleID == "" {
		b.ArticleID = IDFromURL(b.URL)
	}

	b.Title = strings.Join(strings.Fields(b.Title), " ")
	if b.Title == "" {
		return Bookmark{}, ErrArticleTitleRequired
	}
	if utf8.RuneCountInString(b.Title) > MaxTitleRunes {
		return Bookmark{}, fmt.Errorf("タイトルは%d文字以内にしてください: %w", MaxTitleRunes, ErrArticleTitleRequired)
	}

	b.Source = strings.ToLower(strings.TrimSpace(b.Source))
	if utf8.RuneCountInString(b.Source) > MaxSourceLen {
		b.Source = string([]rune(b.Source)[:MaxSourceLen])
	}
	return b, nil
}

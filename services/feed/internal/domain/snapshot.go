package domain

import (
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"
)

const (
	MaxTitleRunes = 300
	MaxSourceLen  = 32
)

type articleSnapshot struct {
	userID    string
	articleID string
	url       string
	title     string
	source    string
}

func normalizeArticleSnapshot(userID, articleID, rawURL, title, source string) (articleSnapshot, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return articleSnapshot{}, ErrUserIDRequired
	}

	rawURL = strings.TrimSpace(rawURL)
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return articleSnapshot{}, ErrInvalidArticleURL
	}

	articleID = strings.TrimSpace(articleID)
	if articleID == "" {
		articleID = IDFromURL(rawURL)
	}

	title = strings.Join(strings.Fields(title), " ")
	if title == "" {
		return articleSnapshot{}, ErrArticleTitleRequired
	}
	if utf8.RuneCountInString(title) > MaxTitleRunes {
		return articleSnapshot{}, fmt.Errorf("タイトルは%d文字以内にしてください: %w", MaxTitleRunes, ErrArticleTitleRequired)
	}

	source = strings.ToLower(strings.TrimSpace(source))
	if utf8.RuneCountInString(source) > MaxSourceLen {
		source = string([]rune(source)[:MaxSourceLen])
	}
	return articleSnapshot{userID: userID, articleID: articleID, url: rawURL, title: title, source: source}, nil
}

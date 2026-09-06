package domain

import "errors"

var (
	ErrUserIDRequired       = errors.New("user id is required")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidTag           = errors.New("invalid tag")
	ErrTagOverlap           = errors.New("同じタグを関心と除外の両方には置けません")
	ErrInvalidArticleURL    = errors.New("記事のURLが不正です")
	ErrArticleTitleRequired = errors.New("記事のタイトルが必要です")
	ErrTooManyBookmarks     = errors.New("ブックマークは200件までです")
	ErrInvalidSearchAfter   = errors.New("不正なページ位置です")
)

func IsValidation(err error) bool {
	return errors.Is(err, ErrUserIDRequired) ||
		errors.Is(err, ErrInvalidTag) ||
		errors.Is(err, ErrTagOverlap) ||
		errors.Is(err, ErrInvalidArticleURL) ||
		errors.Is(err, ErrArticleTitleRequired) ||
		errors.Is(err, ErrTooManyBookmarks) ||
		errors.Is(err, ErrInvalidSearchAfter)
}

package domain

import "errors"

var (
	ErrUserIDRequired = errors.New("user id is required")
	ErrUserNotFound   = errors.New("user not found")
	ErrInvalidTag     = errors.New("invalid tag")
	ErrTagOverlap     = errors.New("同じタグを関心と除外の両方には置けません")
)

func IsValidation(err error) bool {
	return errors.Is(err, ErrUserIDRequired) ||
		errors.Is(err, ErrInvalidTag) ||
		errors.Is(err, ErrTagOverlap)
}

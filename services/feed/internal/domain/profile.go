package domain

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	MaxTags     = 20
	MaxTagRunes = 32
)

type Profile struct {
	UserID       string   `json:"user_id"`
	InterestTags []string `json:"interest_tags"`
	ExcludeTags  []string `json:"exclude_tags"`
}

func EmptyProfile(userID string) Profile {
	return Profile{
		UserID:       userID,
		InterestTags: []string{},
		ExcludeTags:  []string{},
	}
}

func (p Profile) Normalized() (Profile, error) {
	p.UserID = strings.TrimSpace(p.UserID)
	if p.UserID == "" {
		return Profile{}, ErrUserIDRequired
	}

	interest, err := NormalizeTags(p.InterestTags)
	if err != nil {
		return Profile{}, err
	}
	exclude, err := NormalizeTags(p.ExcludeTags)
	if err != nil {
		return Profile{}, err
	}
	if overlap := tagOverlap(interest, exclude); len(overlap) > 0 {
		return Profile{}, fmt.Errorf("%w: %s", ErrTagOverlap, strings.Join(overlap, ", "))
	}

	p.InterestTags = interest
	p.ExcludeTags = exclude
	return p, nil
}

func NormalizeTags(tags []string) ([]string, error) {
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, raw := range tags {
		tag := strings.ToLower(strings.TrimSpace(raw))
		if tag == "" {
			continue
		}
		if utf8.RuneCountInString(tag) > MaxTagRunes {
			return nil, fmt.Errorf("タグは%d文字以内にしてください: %w: %s", MaxTagRunes, ErrInvalidTag, tag)
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	if len(out) > MaxTags {
		return nil, fmt.Errorf("タグは%d個までです: %w", MaxTags, ErrInvalidTag)
	}
	return out, nil
}

func tagOverlap(a, b []string) []string {
	set := make(map[string]struct{}, len(a))
	for _, tag := range a {
		set[tag] = struct{}{}
	}
	var out []string
	for _, tag := range b {
		if _, ok := set[tag]; ok {
			out = append(out, tag)
		}
	}
	return out
}

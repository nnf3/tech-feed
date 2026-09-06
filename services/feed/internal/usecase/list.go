package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

type ListFeed struct {
	index    domain.ArticleIndex
	profiles domain.ProfileStore
	ranker   domain.Ranker
}

func NewListFeed(index domain.ArticleIndex, profiles domain.ProfileStore, ranker domain.Ranker) *ListFeed {
	return &ListFeed{index: index, profiles: profiles, ranker: ranker}
}

func (u *ListFeed) Run(ctx context.Context, query, userID, tag string) ([]domain.Article, error) {
	feedQuery := domain.FeedQuery{Text: query}
	if filter := strings.ToLower(strings.TrimSpace(tag)); filter != "" {
		feedQuery.FilterTags = []string{filter}
	}
	if strings.TrimSpace(userID) != "" {
		profile, err := u.profiles.GetProfile(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("load profile: %w", err)
		}
		if profile != nil {
			feedQuery.InterestTags = profile.InterestTags
			feedQuery.ExcludeTags = profile.ExcludeTags
		}
	}

	items, err := u.index.Search(ctx, feedQuery)
	if err != nil {
		return nil, fmt.Errorf("search articles: %w", err)
	}
	// 関心タグがあるときは ES のスコア順を残す。無いときは新しい順。
	if strings.TrimSpace(query) == "" && len(feedQuery.InterestTags) == 0 && len(feedQuery.FilterTags) == 0 {
		return u.ranker.Rank(items), nil
	}
	return items, nil
}

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
}

func NewListFeed(index domain.ArticleIndex, profiles domain.ProfileStore) *ListFeed {
	return &ListFeed{index: index, profiles: profiles}
}

func (u *ListFeed) Run(ctx context.Context, query, userID, tag, sort, after string) (domain.FeedPage, error) {
	searchAfter, err := domain.DecodeSearchAfter(after)
	if err != nil {
		return domain.FeedPage{}, err
	}

	sort = domain.NormalizeFeedSort(sort)
	feedQuery := domain.FeedQuery{Text: query, Sort: sort, SearchAfter: searchAfter}
	if filter := strings.ToLower(strings.TrimSpace(tag)); filter != "" {
		feedQuery.FilterTags = []string{filter}
	}
	if strings.TrimSpace(userID) != "" {
		profile, err := u.profiles.GetProfile(ctx, userID)
		if err != nil {
			return domain.FeedPage{}, fmt.Errorf("load profile: %w", err)
		}
		if profile != nil {
			feedQuery.ExcludeTags = profile.ExcludeTags
			if sort == domain.SortRecommended {
				feedQuery.InterestTags = profile.InterestTags
			}
		}
	} else {
		feedQuery.Sort = domain.SortNew
	}

	page, err := u.index.Search(ctx, feedQuery)
	if err != nil {
		return domain.FeedPage{}, fmt.Errorf("search articles: %w", err)
	}
	if page.Articles == nil {
		page.Articles = []domain.Article{}
	}
	return page, nil
}

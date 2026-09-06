package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

type ListFeed struct {
	index     domain.ArticleIndex
	profiles  domain.ProfileStore
	bookmarks domain.BookmarkStore
	history   domain.HistoryStore
}

func NewListFeed(index domain.ArticleIndex, profiles domain.ProfileStore, bookmarks domain.BookmarkStore, history domain.HistoryStore) *ListFeed {
	return &ListFeed{index: index, profiles: profiles, bookmarks: bookmarks, history: history}
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
				if err := u.applyActivitySignals(ctx, userID, profile.ExcludeTags, &feedQuery); err != nil {
					return domain.FeedPage{}, err
				}
			}
		} else if sort == domain.SortRecommended {
			if err := u.applyActivitySignals(ctx, userID, nil, &feedQuery); err != nil {
				return domain.FeedPage{}, err
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

// applyActivitySignals は履歴とブックマークからタグと除外 ID を query に載せる。
func (u *ListFeed) applyActivitySignals(ctx context.Context, userID string, excludeTags []string, query *domain.FeedQuery) error {
	var bookmarkIDs []string
	var bookmarks []domain.Bookmark
	if u.bookmarks != nil {
		items, err := u.bookmarks.ListBookmarks(ctx, userID)
		if err != nil {
			return fmt.Errorf("load bookmarks: %w", err)
		}
		bookmarks = items
		for _, item := range items {
			bookmarkIDs = append(bookmarkIDs, item.ArticleID)
		}
	}

	var historyIDs []string
	var recent []domain.HistoryEntry
	if u.history != nil {
		items, err := u.history.ListHistory(ctx, userID)
		if err != nil {
			return fmt.Errorf("load history: %w", err)
		}
		for _, item := range items {
			historyIDs = append(historyIDs, item.ArticleID)
		}
		if len(items) > domain.RecommendHistoryLookback {
			items = items[:domain.RecommendHistoryLookback]
		}
		recent = items
	}

	query.ExcludeIDs = domain.UniqueIDs(bookmarkIDs, historyIDs)
	var recentIDs []string
	for _, item := range recent {
		recentIDs = append(recentIDs, item.ArticleID)
	}
	lookup := domain.UniqueIDs(bookmarkIDs, recentIDs)
	if len(lookup) == 0 {
		return nil
	}

	articles, err := u.index.GetByIDs(ctx, lookup)
	if err != nil {
		return fmt.Errorf("load activity articles: %w", err)
	}
	byID := domain.ArticlesByID(articles)
	query.BookmarkTags = domain.TopTags(domain.BookmarkTagWeights(byID, bookmarks), domain.RecommendTagLimit, excludeTags)
	query.HistoryTags = domain.TopTags(domain.HistoryTagWeights(byID, recent), domain.RecommendTagLimit, excludeTags)
	return nil
}

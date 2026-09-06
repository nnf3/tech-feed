package usecase

import (
	"context"
	"strings"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

type Bookmarks struct {
	users domain.UserStore
	store domain.BookmarkStore
}

func NewBookmarks(users domain.UserStore, store domain.BookmarkStore) *Bookmarks {
	return &Bookmarks{users: users, store: store}
}

func (b *Bookmarks) requireUser(ctx context.Context, userID string) (string, error) {
	user, err := domain.User{ID: userID}.Normalized()
	if err != nil {
		return "", err
	}
	found, err := b.users.Get(ctx, user.ID)
	if err != nil {
		return "", err
	}
	if found == nil {
		return "", domain.ErrUserNotFound
	}
	return user.ID, nil
}

func (b *Bookmarks) List(ctx context.Context, userID string) ([]domain.Bookmark, error) {
	id, err := b.requireUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	items, err := b.store.ListBookmarks(ctx, id)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return []domain.Bookmark{}, nil
	}
	return items, nil
}

func (b *Bookmarks) Add(ctx context.Context, bookmark domain.Bookmark) (domain.Bookmark, error) {
	bookmark, err := bookmark.Normalized()
	if err != nil {
		return domain.Bookmark{}, err
	}
	if _, err := b.requireUser(ctx, bookmark.UserID); err != nil {
		return domain.Bookmark{}, err
	}

	existing, err := b.store.GetBookmark(ctx, bookmark.UserID, bookmark.ArticleID)
	if err != nil {
		return domain.Bookmark{}, err
	}
	if existing == nil {
		n, err := b.store.CountBookmarks(ctx, bookmark.UserID)
		if err != nil {
			return domain.Bookmark{}, err
		}
		if n >= domain.MaxBookmarks {
			return domain.Bookmark{}, domain.ErrTooManyBookmarks
		}
	}

	if err := b.store.UpsertBookmark(ctx, bookmark); err != nil {
		return domain.Bookmark{}, err
	}
	saved, err := b.store.GetBookmark(ctx, bookmark.UserID, bookmark.ArticleID)
	if err != nil {
		return domain.Bookmark{}, err
	}
	if saved == nil {
		return bookmark, nil
	}
	return *saved, nil
}

func (b *Bookmarks) Remove(ctx context.Context, userID, articleID string) error {
	id, err := b.requireUser(ctx, userID)
	if err != nil {
		return err
	}
	articleID = strings.TrimSpace(articleID)
	if articleID == "" {
		return domain.ErrInvalidArticleURL
	}
	return b.store.DeleteBookmark(ctx, id, articleID)
}

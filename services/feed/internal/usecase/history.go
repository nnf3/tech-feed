package usecase

import (
	"context"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

type History struct {
	users domain.UserStore
	store domain.HistoryStore
}

func NewHistory(users domain.UserStore, store domain.HistoryStore) *History {
	return &History{users: users, store: store}
}

func (h *History) requireUser(ctx context.Context, userID string) (string, error) {
	user, err := domain.User{ID: userID}.Normalized()
	if err != nil {
		return "", err
	}
	found, err := h.users.Get(ctx, user.ID)
	if err != nil {
		return "", err
	}
	if found == nil {
		return "", domain.ErrUserNotFound
	}
	return user.ID, nil
}

func (h *History) List(ctx context.Context, userID string) ([]domain.HistoryEntry, error) {
	id, err := h.requireUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	items, err := h.store.ListHistory(ctx, id)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return []domain.HistoryEntry{}, nil
	}
	return items, nil
}

func (h *History) Record(ctx context.Context, entry domain.HistoryEntry) (domain.HistoryEntry, error) {
	entry, err := entry.Normalized()
	if err != nil {
		return domain.HistoryEntry{}, err
	}
	if _, err := h.requireUser(ctx, entry.UserID); err != nil {
		return domain.HistoryEntry{}, err
	}
	if err := h.store.UpsertHistory(ctx, entry); err != nil {
		return domain.HistoryEntry{}, err
	}
	if err := h.store.TrimHistory(ctx, entry.UserID, domain.MaxHistory); err != nil {
		return domain.HistoryEntry{}, err
	}
	saved, err := h.store.GetHistory(ctx, entry.UserID, entry.ArticleID)
	if err != nil {
		return domain.HistoryEntry{}, err
	}
	if saved == nil {
		return entry, nil
	}
	return *saved, nil
}

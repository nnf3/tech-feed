package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

type Users struct {
	store domain.UserStore
}

func NewUsers(store domain.UserStore) *Users {
	return &Users{store: store}
}

func (u *Users) Upsert(ctx context.Context, user domain.User) (domain.User, error) {
	user.ID = strings.TrimSpace(user.ID)
	if user.ID == "" {
		return domain.User{}, fmt.Errorf("user id is required")
	}
	if err := u.store.Upsert(ctx, user); err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (u *Users) Get(ctx context.Context, id string) (*domain.User, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("user id is required")
	}
	return u.store.Get(ctx, id)
}

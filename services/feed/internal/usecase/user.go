package usecase

import (
	"context"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

type Users struct {
	store domain.UserStore
}

func NewUsers(store domain.UserStore) *Users {
	return &Users{store: store}
}

func (u *Users) Upsert(ctx context.Context, user domain.User) (domain.User, error) {
	user, err := user.Normalized()
	if err != nil {
		return domain.User{}, err
	}
	if err := u.store.Upsert(ctx, user); err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (u *Users) Get(ctx context.Context, id string) (*domain.User, error) {
	user, err := domain.User{ID: id}.Normalized()
	if err != nil {
		return nil, err
	}
	return u.store.Get(ctx, user.ID)
}

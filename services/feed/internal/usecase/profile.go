package usecase

import (
	"context"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

type Profiles struct {
	users domain.UserStore
	store domain.ProfileStore
}

func NewProfiles(users domain.UserStore, store domain.ProfileStore) *Profiles {
	return &Profiles{users: users, store: store}
}

func (p *Profiles) Get(ctx context.Context, userID string) (domain.Profile, error) {
	user, err := domain.User{ID: userID}.Normalized()
	if err != nil {
		return domain.Profile{}, err
	}
	profile, err := p.store.GetProfile(ctx, user.ID)
	if err != nil {
		return domain.Profile{}, err
	}
	if profile == nil {
		return domain.EmptyProfile(user.ID), nil
	}
	return *profile, nil
}

func (p *Profiles) Upsert(ctx context.Context, profile domain.Profile) (domain.Profile, error) {
	profile, err := profile.Normalized()
	if err != nil {
		return domain.Profile{}, err
	}

	user, err := p.users.Get(ctx, profile.UserID)
	if err != nil {
		return domain.Profile{}, err
	}
	if user == nil {
		return domain.Profile{}, domain.ErrUserNotFound
	}

	if err := p.store.UpsertProfile(ctx, profile); err != nil {
		return domain.Profile{}, err
	}
	return profile, nil
}

package postgres

import (
	"context"
	"errors"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *Store) UpsertProfile(ctx context.Context, profile domain.Profile) error {
	row := profileFromDomain(profile)
	return s.withCtx(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"interest_tags", "exclude_tags", "updated_at"}),
	}).Create(&row).Error
}

func (s *Store) GetProfile(ctx context.Context, userID string) (*domain.Profile, error) {
	var row profileRecord
	err := s.withCtx(ctx).First(&row, "user_id = ?", userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	profile := row.toDomain()
	return &profile, nil
}

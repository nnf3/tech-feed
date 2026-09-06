package postgres

import (
	"context"
	"errors"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *Store) Upsert(ctx context.Context, user domain.User) error {
	row := userFromDomain(user)
	return s.withCtx(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"email", "name", "updated_at"}),
	}).Create(&row).Error
}

func (s *Store) Get(ctx context.Context, id string) (*domain.User, error) {
	var row userRecord
	err := s.withCtx(ctx).First(&row, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	user := row.toDomain()
	return &user, nil
}

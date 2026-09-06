package postgres

import (
	"context"
	"errors"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *Store) UpsertBookmark(ctx context.Context, bookmark domain.Bookmark) error {
	row := bookmarkFromDomain(bookmark)
	return s.withCtx(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "article_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"url", "title", "source"}),
	}).Create(&row).Error
}

func (s *Store) GetBookmark(ctx context.Context, userID, articleID string) (*domain.Bookmark, error) {
	var row bookmarkRecord
	err := s.withCtx(ctx).First(&row, "user_id = ? AND article_id = ?", userID, articleID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	item := row.toDomain()
	return &item, nil
}

func (s *Store) DeleteBookmark(ctx context.Context, userID, articleID string) error {
	return s.withCtx(ctx).Where("user_id = ? AND article_id = ?", userID, articleID).Delete(&bookmarkRecord{}).Error
}

func (s *Store) ListBookmarks(ctx context.Context, userID string) ([]domain.Bookmark, error) {
	var rows []bookmarkRecord
	if err := s.withCtx(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]domain.Bookmark, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.toDomain())
	}
	return items, nil
}

func (s *Store) CountBookmarks(ctx context.Context, userID string) (int, error) {
	var n int64
	if err := s.withCtx(ctx).Model(&bookmarkRecord{}).Where("user_id = ?", userID).Count(&n).Error; err != nil {
		return 0, err
	}
	return int(n), nil
}

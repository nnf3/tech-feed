package postgres

import (
	"context"
	"errors"

	"github.com/nnf3/tech-feed/services/feed/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *Store) UpsertHistory(ctx context.Context, entry domain.HistoryEntry) error {
	row := historyFromDomain(entry)
	return s.withCtx(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "article_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"url":        gorm.Expr("EXCLUDED.url"),
			"title":      gorm.Expr("EXCLUDED.title"),
			"source":     gorm.Expr("EXCLUDED.source"),
			"viewed_at":  gorm.Expr("now()"),
			"view_count": gorm.Expr("article_views.view_count + 1"),
		}),
	}).Create(&row).Error
}

func (s *Store) GetHistory(ctx context.Context, userID, articleID string) (*domain.HistoryEntry, error) {
	var row historyRecord
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

func (s *Store) ListHistory(ctx context.Context, userID string) ([]domain.HistoryEntry, error) {
	var rows []historyRecord
	if err := s.withCtx(ctx).Where("user_id = ?", userID).Order("viewed_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]domain.HistoryEntry, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.toDomain())
	}
	return items, nil
}

func (s *Store) TrimHistory(ctx context.Context, userID string, keep int) error {
	if keep <= 0 {
		return nil
	}
	return s.withCtx(ctx).Exec(`
		DELETE FROM article_views
		WHERE user_id = ?
		  AND article_id NOT IN (
		    SELECT article_id FROM article_views
		    WHERE user_id = ?
		    ORDER BY viewed_at DESC
		    LIMIT ?
		  )
	`, userID, userID, keep).Error
}

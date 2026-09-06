package postgres

import (
	"context"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Store struct {
	db *gorm.DB
}

func New(ctx context.Context, databaseURL string) (*Store, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		// スキーマは migrations/*.sql が正。AutoMigrate は使わない。
		DisableAutomaticPing: true,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() {
	sqlDB, err := s.db.DB()
	if err != nil {
		return
	}
	_ = sqlDB.Close()
}

func (s *Store) withCtx(ctx context.Context) *gorm.DB {
	return s.db.WithContext(ctx)
}

package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nnf3/tech-feed/services/feed/internal/domain"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, databaseURL string) (*Store, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) Upsert(ctx context.Context, user domain.User) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO users (id, email, name)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET
			email = EXCLUDED.email,
			name = EXCLUDED.name,
			updated_at = now()
	`, user.ID, user.Email, user.Name)
	return err
}

func (s *Store) Get(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, name FROM users WHERE id = $1
	`, id).Scan(&user.ID, &user.Email, &user.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

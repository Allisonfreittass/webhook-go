package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) CreateEvent(ctx context.Context, source string, headers, body []byte) (string, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var eventID string
	err = tx.QueryRow(ctx,
		`INSERT INTO events (source, headers, body) VALUES ($1, $2, $3) RETURNING id`, source, headers, body,
	).Scan(&eventID)
	if err != nil {
		return "", fmt.Errorf("insert event: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO deliveries (event_id) VALUES ($1)`, eventID,
	)
	if err != nil {
		return "", fmt.Errorf("insert delivery: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("error in commit: %w", err)
	}

	return eventID, nil
}

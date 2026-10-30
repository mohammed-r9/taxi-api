package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store wraps the connection pool and is the place to hang generated
// queries on. After running `sqlc generate`, embed *Queries here.
type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Begin(ctx context.Context) (pgx.Tx, error) {
	return s.pool.Begin(ctx)
}

func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

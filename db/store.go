package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store wraps the connection pool and the generated sqlc queries.
type Store struct {
	*Queries
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		Queries: New(pool),
		pool:    pool,
	}
}

func (s *Store) Begin(ctx context.Context) (pgx.Tx, error) {
	return s.pool.Begin(ctx)
}

func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

// IsUniqueViolation reports whether err is a Postgres unique constraint
// violation.
func IsUniqueViolation(err error) bool {
	return isPgError(err, "23505")
}

// IsForeignKeyViolation reports whether err is a Postgres foreign key
// violation.
func IsForeignKeyViolation(err error) bool {
	return isPgError(err, "23503")
}

func isPgError(err error, code string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code
}

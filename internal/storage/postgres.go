package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgresStorage implements Storage using a PostgreSQL database.
type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage opens a connection pool to PostgreSQL, verifies
// connectivity, and returns a ready-to-use PostgresStorage.
func NewPostgresStorage(dsn string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	// Tune the connection pool for a long-running service.
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("postgres ping: %w", err)
	}

	return &PostgresStorage{db: db}, nil
}

// Close releases the connection pool.
func (s *PostgresStorage) Close() error {
	return s.db.Close()
}

// SaveBlurredImage inserts a blurred image record.
func (s *PostgresStorage) SaveBlurredImage(ctx context.Context, sourceID, blurredPath string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO blurred_img (source_img_id, blurred_img_url) VALUES ($1, $2)`,
		sourceID, blurredPath,
	)
	if err != nil {
		return fmt.Errorf("insert blurred_img: %w", err)
	}
	return nil
}

// DeleteBlurredImage deletes all records for sourceID.
func (s *PostgresStorage) DeleteBlurredImage(ctx context.Context, sourceID string) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM blurred_img WHERE source_img_id = $1`,
		sourceID,
	)
	if err != nil {
		return 0, fmt.Errorf("delete blurred_img: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected: %w", err)
	}
	return rows, nil
}

// Compile-time check that PostgresStorage satisfies Storage.
var _ Storage = (*PostgresStorage)(nil)

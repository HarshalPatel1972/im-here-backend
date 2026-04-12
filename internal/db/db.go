package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database url: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	if err := runMigrations(ctx, pool); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return pool, nil
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrationsPath := filepath.Join("internal", "db", "migrations", "001_init.sql")
	content, err := os.ReadFile(migrationsPath)
	if err != nil {
		return fmt.Errorf("unable to read migration file: %w", err)
	}

	_, err = pool.Exec(ctx, string(content))
	return err
}

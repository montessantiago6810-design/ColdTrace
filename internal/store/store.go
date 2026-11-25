// Package store owns the PostgreSQL schema (goose migrations embedded in the
// binary) and opens connection pools for the sqlc queries in store/db.
package store

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the embedded migration files (for tooling and tests).
func Migrations() fs.FS {
	sub, err := fs.Sub(migrationFiles, "migrations")
	if err != nil {
		panic(err) // the directory is embedded at compile time
	}
	return sub
}

// OpenPool connects to databaseURL and checks the connection.
func OpenPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		// pgx errors can include the connection string; keep it out of logs.
		return nil, errors.New("parse database URL: invalid connection string")
	}
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.ConnConfig.RuntimeParams["application_name"] = "coldtrace"

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("open pool: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

// Migrator applies the embedded migrations through a pool.
type Migrator struct {
	provider *goose.Provider
	db       *sql.DB
}

// NewMigrator wraps pool for goose. Close releases the database/sql handle,
// not the pool.
func NewMigrator(pool *pgxpool.Pool, logger *slog.Logger) (*Migrator, error) {
	db := stdlib.OpenDBFromPool(pool)
	opts := []goose.ProviderOption{}
	if logger != nil {
		opts = append(opts, goose.WithLogger(slogAdapter{logger}), goose.WithVerbose(true))
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, db, Migrations(), opts...)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("load migrations: %w", err)
	}
	return &Migrator{provider: provider, db: db}, nil
}

// Up applies every pending migration and returns the resulting version.
func (m *Migrator) Up(ctx context.Context) (int64, error) {
	if _, err := m.provider.Up(ctx); err != nil {
		return 0, fmt.Errorf("migrate up: %w", err)
	}
	return m.provider.GetDBVersion(ctx)
}

// DownTo rolls back to version (0 removes the whole schema).
func (m *Migrator) DownTo(ctx context.Context, version int64) (int64, error) {
	if _, err := m.provider.DownTo(ctx, version); err != nil {
		return 0, fmt.Errorf("migrate down to %d: %w", version, err)
	}
	return m.provider.GetDBVersion(ctx)
}

// Status lists each migration with whether it has been applied.
func (m *Migrator) Status(ctx context.Context) ([]*goose.MigrationStatus, error) {
	return m.provider.Status(ctx)
}

// Close releases the database/sql handle.
func (m *Migrator) Close() error {
	return m.db.Close()
}

type slogAdapter struct{ logger *slog.Logger }

func (a slogAdapter) Printf(format string, v ...any) {
	a.logger.Info(fmt.Sprintf(format, v...), slog.String("component", "migrate"))
}

func (a slogAdapter) Fatalf(format string, v ...any) {
	a.logger.Error(fmt.Sprintf(format, v...), slog.String("component", "migrate"))
}

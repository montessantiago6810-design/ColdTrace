//go:build integration

package store_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/montessantiago6810-design/coldtrace/internal/store"
	"github.com/montessantiago6810-design/coldtrace/internal/store/db"
)

// timescaleImage matches the image in compose.yaml.
const timescaleImage = "timescale/timescaledb:2.32.1-pg17"

// databaseURL returns COLDTRACE_TEST_DATABASE_URL when set (an empty,
// disposable database), otherwise starts a TimescaleDB container.
func databaseURL(t *testing.T) string {
	t.Helper()
	if dsn := os.Getenv("COLDTRACE_TEST_DATABASE_URL"); dsn != "" {
		return dsn
	}
	ctx := context.Background()
	ctr, err := tcpostgres.Run(ctx, timescaleImage,
		tcpostgres.WithDatabase("coldtrace_test"),
		tcpostgres.WithUsername("coldtrace"),
		tcpostgres.WithPassword("coldtrace"),
		tcpostgres.BasicWaitStrategies(),
	)
	testcontainers.CleanupContainer(t, ctr)
	if err != nil {
		t.Fatalf("start TimescaleDB container: %v", err)
	}
	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}
	return dsn
}

func migratedPool(t *testing.T) (*pgxpool.Pool, *store.Migrator) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool, err := store.OpenPool(ctx, databaseURL(t))
	if err != nil {
		t.Fatalf("OpenPool: %v", err)
	}
	t.Cleanup(pool.Close)

	m, err := store.NewMigrator(pool, nil)
	if err != nil {
		t.Fatalf("NewMigrator: %v", err)
	}
	t.Cleanup(func() { _ = m.Close() })

	if _, err := m.DownTo(ctx, 0); err != nil {
		t.Fatalf("reset schema: %v", err)
	}
	if v, err := m.Up(ctx); err != nil || v != 1 {
		t.Fatalf("Up = %d, %v; want version 1", v, err)
	}
	return pool, m
}

func TestOrganizationQueries(t *testing.T) {
	pool, _ := migratedPool(t)
	ctx := context.Background()
	q := db.New(pool)

	created, err := q.CreateOrganization(ctx, db.CreateOrganizationParams{Slug: "nordic-pharma", Name: "Nordic Pharma Logistics"})
	if err != nil {
		t.Fatalf("CreateOrganization: %v", err)
	}
	if created.ID.String() == "" || created.CreatedAt.IsZero() {
		t.Fatalf("defaults not filled: %+v", created)
	}

	got, err := q.GetOrganizationBySlug(ctx, "nordic-pharma")
	if err != nil {
		t.Fatalf("GetOrganizationBySlug: %v", err)
	}
	if got.ID != created.ID || got.Name != "Nordic Pharma Logistics" {
		t.Fatalf("got %+v, want %+v", got, created)
	}

	if _, err := q.GetOrganizationBySlug(ctx, "missing"); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("missing slug: err = %v, want pgx.ErrNoRows", err)
	}

	if _, err := q.CreateOrganization(ctx, db.CreateOrganizationParams{Slug: "fresh-fleet", Name: "Fresh Fleet"}); err != nil {
		t.Fatalf("second org: %v", err)
	}
	list, err := q.ListOrganizations(ctx, 10)
	if err != nil {
		t.Fatalf("ListOrganizations: %v", err)
	}
	if len(list) != 2 || list[0].Slug != "fresh-fleet" || list[1].Slug != "nordic-pharma" {
		t.Fatalf("list not ordered by slug: %+v", list)
	}
	if limited, _ := q.ListOrganizations(ctx, 1); len(limited) != 1 {
		t.Fatalf("limit ignored: %d rows", len(limited))
	}
}

func TestOrganizationConstraints(t *testing.T) {
	pool, _ := migratedPool(t)
	ctx := context.Background()
	q := db.New(pool)

	if _, err := q.CreateOrganization(ctx, db.CreateOrganizationParams{Slug: "acme", Name: "Acme"}); err != nil {
		t.Fatalf("CreateOrganization: %v", err)
	}

	cases := []struct {
		name, slug, orgName, constraint string
	}{
		{"duplicate slug", "acme", "Acme again", "organizations_slug_key"},
		{"uppercase slug", "Acme-2", "Acme 2", "organizations_slug_format"},
		{"trailing hyphen", "acme-", "Acme 3", "organizations_slug_format"},
		{"blank name", "blank", "   ", "organizations_name_length"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := q.CreateOrganization(ctx, db.CreateOrganizationParams{Slug: tc.slug, Name: tc.orgName})
			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) || pgErr.ConstraintName != tc.constraint {
				t.Fatalf("err = %v, want violation of %s", err, tc.constraint)
			}
		})
	}
}

func TestMigrationsRollBackCleanly(t *testing.T) {
	pool, m := migratedPool(t)
	ctx := context.Background()

	if v, err := m.DownTo(ctx, 0); err != nil || v != 0 {
		t.Fatalf("DownTo(0) = %d, %v", v, err)
	}
	var exists bool
	if err := pool.QueryRow(ctx, "SELECT to_regclass('public.organizations') IS NOT NULL").Scan(&exists); err != nil {
		t.Fatalf("check table: %v", err)
	}
	if exists {
		t.Fatal("organizations still exists after rolling back to 0")
	}
	if v, err := m.Up(ctx); err != nil || v != 1 {
		t.Fatalf("Up after rollback = %d, %v", v, err)
	}
}

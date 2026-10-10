// Package testdb provides a shared helper for integration tests that need a
// real database. It connects to the test database, applies pending migrations,
// and truncates all tables so each test starts from a clean state. A Factory
// (see NewFactory) then inserts test data on that pool.
//
// Usage:
//
//	func TestSomething(t *testing.T) {
//	    pool := testdb.New(t, "../adapters/postgresql/migrations")
//	    f := testdb.NewFactory(t, pool)
//	    product := f.Product(testdb.WithName("widget"))
//	    q := repo.New(pool)
//	    // ...
//	}
//
// The DSN is read from TEST_DB_DSN (real env vars win), then from the .env.test
// file at the repo root, and finally falls back to a local ecom_test default.
// Tests are skipped in -short mode or when the database is unreachable, so
// `go test ./...` stays green without infrastructure running.
package testdb

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver for goose
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
)

// defaultDSN points at the local ecom_test database inside the ecom-postgres
// Docker container. Override with TEST_DB_DSN for CI or remote databases.
const defaultDSN = "host=localhost user=postgres password=postgres dbname=ecom_test sslmode=disable"

// New returns a pgxpool connected to the test database with migrations applied
// and all tables truncated. The pool is closed automatically when the test
// finishes. migrationsDir is resolved relative to the calling test's package
// directory (e.g. "../adapters/postgresql/migrations").
func New(t *testing.T, migrationsDir string) *pgxpool.Pool {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}

	// Load .env.test from the repo root if present. Real environment
	// variables always win — godotenv.Load never overwrites existing vars.
	loadDotenv(t)

	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = defaultDSN
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	// Safety guard: never mutate a database that doesn't look like a test one.
	// This protects local data from a misconfigured TEST_DB_DSN.
	var dbName string
	probe, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skipf("test database unreachable (%v) — set TEST_DB_DSN or start the DB", err)
	}
	if err := probe.QueryRow(ctx, "SELECT current_database()").Scan(&dbName); err != nil {
		probe.Close()
		t.Fatalf("probe test database: %v", err)
	}
	probe.Close()
	if !strings.Contains(dbName, "test") {
		t.Fatalf("refusing to run tests against database %q — TEST_DB_DSN must point at a database with 'test' in its name", dbName)
	}

	migrate(t, ctx, dsn, migrationsDir)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect pool to test db: %v", err)
	}
	t.Cleanup(pool.Close)

	truncate(t, ctx, pool)

	return pool
}

// loadDotenv finds .env.test by walking up from the working directory (tests
// run from their own package dir) and loads it. Missing file is fine — the
// default DSN and real env vars still apply.
func loadDotenv(t *testing.T) {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		return
	}
	for {
		candidate := filepath.Join(dir, ".env.test")
		if _, err := os.Stat(candidate); err == nil {
			// Errors are non-fatal: malformed lines or unreadable files should
			// not fail every test — the fallback DSN still works.
			_ = godotenv.Load(candidate)
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return // reached filesystem root without finding the file
		}
		dir = parent
	}
}

// migrate applies pending goose migrations using the same migration files as
// the application, so the test schema always tracks the source of truth.
func migrate(t *testing.T, ctx context.Context, dsn, migrationsDir string) {
	t.Helper()

	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db for migrations: %v", err)
	}
	defer sqlDB.Close()

	goose.SetLogger(goose.NopLogger())
	if err := goose.UpContext(ctx, sqlDB, migrationsDir); err != nil {
		t.Fatalf("apply migrations from %s: %v", migrationsDir, err)
	}
}

// truncate empties every public table (except goose's bookkeeping table) so
// tests don't inherit rows from previous runs. Tables are discovered from the
// catalog rather than hardcoded, so new tables are covered automatically.
func truncate(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	rows, err := pool.Query(ctx, `
		SELECT tablename FROM pg_tables
		WHERE schemaname = 'public' AND tablename <> 'goose_db_version'`)
	if err != nil {
		t.Fatalf("list tables: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan table name: %v", err)
		}
		tables = append(tables, fmt.Sprintf("public.%s", name))
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate tables: %v", err)
	}
	if len(tables) == 0 {
		return
	}

	stmt := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", strings.Join(tables, ", "))
	if _, err := pool.Exec(ctx, stmt); err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}

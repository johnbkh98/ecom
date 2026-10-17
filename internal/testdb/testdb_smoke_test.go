package testdb

import (
	"testing"
)

func TestNewReturnsUsablePool(t *testing.T) {
	pool := New(t, "../adapters/postgresql/migrations")

	var dbName string
	if err := pool.QueryRow(t.Context(), "SELECT current_database()").Scan(&dbName); err != nil {
		t.Fatalf("query: %v", err)
	}
	if dbName != "ecom_test" {
		t.Fatalf("connected to %q, want ecom_test", dbName)
	}

	// seed a row, then confirm a second New() starts clean (truncate works)
	_, err := pool.Exec(t.Context(), "INSERT INTO products (name, price, quantity) VALUES ('smoke', 100, 5)")
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	pool2 := New(t, "../adapters/postgresql/migrations")
	var n int
	if err := pool2.QueryRow(t.Context(), "SELECT count(*) FROM products").Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0 products after truncate, got %d", n)
	}
}

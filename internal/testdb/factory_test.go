package testdb

import "testing"

func TestFactoryCreatesProducts(t *testing.T) {
	pool := New(t, "../adapters/postgresql/migrations")
	f := NewFactory(t, pool)

	// overrides are applied
	p := f.Product(WithName("widget"), WithPrice(999), WithQuantity(3))
	if p.ID == 0 {
		t.Fatal("expected DB-assigned id, got 0")
	}
	if p.Name != "widget" || p.Price != 999 || p.Quantity != 3 {
		t.Errorf("overrides not applied: %+v", p)
	}
	if p.CreatedAt.Time.IsZero() {
		t.Error("expected non-zero created_at")
	}

	// defaults: unique name, price 1000, quantity 10
	d := f.Product()
	if d.Name == p.Name {
		t.Errorf("default product names should be unique, got %q twice", d.Name)
	}
	if d.Price != 1000 || d.Quantity != 10 {
		t.Errorf("defaults = price %d quantity %d, want 1000/10", d.Price, d.Quantity)
	}

	// bulk creation
	if got := f.Products(3); len(got) != 3 {
		t.Fatalf("Products(3) returned %d rows", len(got))
	}

	var n int
	if err := pool.QueryRow(t.Context(), "SELECT count(*) FROM products").Scan(&n); err != nil {
		t.Fatalf("count products: %v", err)
	}
	if n != 5 {
		t.Errorf("products in db = %d, want 5", n)
	}
}

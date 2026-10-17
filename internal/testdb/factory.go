package testdb

import (
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	repo "github.com/johnbkh98/ecom/internal/adapters/postgresql/sqlc"
)

// Factory inserts test data into the test database. Build it on a pool
// returned by New (so migrations have already been applied and tables
// truncated), then call its methods to create rows.
//
// Usage:
//
//	pool := testdb.New(t, "../adapters/postgresql/migrations")
//	f := testdb.NewFactory(t, pool)
//	product := f.Product(testdb.WithName("widget"), testdb.WithPrice(999))
type Factory struct {
	t   *testing.T
	q   repo.Querier
	seq int
}

// NewFactory binds a factory to an existing test database pool. It never
// migrates or truncates — those belong to New — so factories can't wipe data
// a test has already seeded.
func NewFactory(t *testing.T, pool *pgxpool.Pool) *Factory {
	t.Helper()
	return &Factory{t: t, q: repo.New(pool)}
}

// ProductOption mutates a product before it is inserted.
type ProductOption func(*repo.Product)

// WithName overrides the product name.
func WithName(name string) ProductOption {
	return func(p *repo.Product) { p.Name = name }
}

// WithPrice overrides the product price (the schema enforces price >= 0).
func WithPrice(price int32) ProductOption {
	return func(p *repo.Product) { p.Price = price }
}

// WithQuantity overrides the product quantity.
func WithQuantity(quantity int32) ProductOption {
	return func(p *repo.Product) { p.Quantity = quantity }
}

// Product inserts a product with test-friendly defaults and returns the row
// as stored, including the DB-assigned id and DB-generated created_at
// (DEFAULT now()). Defaults: a unique name ("product-N", so bulk-created rows
// never collide), price 1000 and quantity 10. The test fails on any insert
// error.
func (f *Factory) Product(opts ...ProductOption) repo.Product {
	f.t.Helper()

	f.seq++
	p := repo.Product{
		Name:     fmt.Sprintf("product-%d", f.seq),
		Price:    1000,
		Quantity: 10,
	}
	for _, opt := range opts {
		opt(&p)
	}

	stored, err := f.q.CreateProduct(f.t.Context(), repo.CreateProductParams{
		Name:     p.Name,
		Quantity: p.Quantity,
		Price:    p.Price,
	})
	if err != nil {
		f.t.Fatalf("factory: insert product %q: %v", p.Name, err)
	}
	return stored
}

// Products inserts n products with default attributes and returns the stored
// rows. Handy for padding out list endpoints.
func (f *Factory) Products(n int) []repo.Product {
	f.t.Helper()

	products := make([]repo.Product, 0, n)
	for i := 0; i < n; i++ {
		products = append(products, f.Product())
	}
	return products
}

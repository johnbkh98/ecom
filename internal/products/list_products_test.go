package products

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	repo "github.com/johnbkh98/ecom/internal/adapters/postgresql/sqlc"
	"github.com/johnbkh98/ecom/internal/testdb"
)

// TestMain is a placeholder for package-level test setup (e.g. a shared test
// database). Note: m.Run() must be called — without it the test binary exits
// immediately and no tests in this package ever execute.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// doListRequest exercises the handler directly. ListProducts has no URL
// params, so no chi router is needed.
func doListRequest(h *handler) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rr := httptest.NewRecorder()
	h.ListProducts(rr, req)
	return rr
}

// fakeQuerier implements repo.Querier so the handler can be tested without a
// database. Only ListProducts returns meaningful values; the rest are no-ops.
type fakeQuerier struct {
	products []repo.Product
	err      error
}

func (f *fakeQuerier) ListProducts(context.Context) ([]repo.Product, error) {
	return f.products, f.err
}

func (f *fakeQuerier) FindProductById(context.Context, int64) (repo.Product, error) {
	return repo.Product{}, nil
}

func (f *fakeQuerier) CreateProduct(context.Context, repo.CreateProductParams) (repo.Product, error) {
	return repo.Product{}, nil
}

func (f *fakeQuerier) CreateOrder(context.Context, int64) (repo.Order, error) {
	return repo.Order{}, nil
}

func (f *fakeQuerier) CreateOrderItem(context.Context, repo.CreateOrderItemParams) (repo.OrderItem, error) {
	return repo.OrderItem{}, nil
}

// --- Unit tests (no database, run fine with -short) ---

func TestListProductsReturnsProducts(t *testing.T) {
	want := []repo.Product{
		{ID: 1, Name: "widget", Price: 999, Quantity: 3},
		{ID: 2, Name: "gadget", Price: 1500, Quantity: 10},
	}
	h := NewHandler(NewService(&fakeQuerier{products: want}))

	rr := doListRequest(h)

	assertListResponse(t, rr, want)
}

func TestListProductsEmptyTableReturnsEmptyArray(t *testing.T) {
	h := NewHandler(NewService(&fakeQuerier{})) // nil products, nil error

	rr := doListRequest(h)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	// The handler must coerce nil → []; a raw nil slice would encode as `null`.
	if body := strings.TrimSpace(rr.Body.String()); body != "[]" {
		t.Fatalf("body = %q, want []", body)
	}
}

func TestListProductsRepoErrorReturns500(t *testing.T) {
	h := NewHandler(NewService(&fakeQuerier{err: errors.New("db down")}))

	rr := doListRequest(h)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(rr.Body.String(), "db down") {
		t.Fatalf("body = %q, want it to contain the repo error", rr.Body.String())
	}
}

// --- Integration tests (real Postgres via testdb helper) ---

func TestListProductsIntegration(t *testing.T) {
	pool := testdb.New(t, "../adapters/postgresql/migrations")
	f := testdb.NewFactory(t, pool)
	h := NewHandler(NewService(repo.New(pool)))

	// create two products; IDs are DB-assigned, so the factory returns the stored rows
	seeded := []repo.Product{
		f.Product(testdb.WithPrice(999), testdb.WithQuantity(3)),
		f.Product(testdb.WithName("gadget"), testdb.WithPrice(1500), testdb.WithQuantity(10)),
	}

	rr := doListRequest(h)

	assertListResponse(t, rr, seeded)
}

func TestListProductsIntegrationEmptyDB(t *testing.T) {
	pool := testdb.New(t, "../adapters/postgresql/migrations") // truncates tables
	h := NewHandler(NewService(repo.New(pool)))

	rr := doListRequest(h)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if body := strings.TrimSpace(rr.Body.String()); body != "[]" {
		t.Fatalf("body = %q, want []", body)
	}
}

// assertListResponse checks the shared response contract: 200, JSON content
// type, and a body that decodes to exactly the wanted products.
func assertListResponse(t *testing.T, rr *httptest.ResponseRecorder, want []repo.Product) {
	t.Helper()

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}

	var got []repo.Product
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d products, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].ID != want[i].ID || got[i].Name != want[i].Name ||
			got[i].Price != want[i].Price || got[i].Quantity != want[i].Quantity {
			t.Errorf("product[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

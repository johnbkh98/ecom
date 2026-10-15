package products

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	repo "github.com/johnbkh98/ecom/internal/adapters/postgresql/sqlc"
	"github.com/johnbkh98/ecom/internal/testdb"
)

// TestMain is a placeholder for package-level test setup (e.g. a shared test
// database). Note: m.Run() must be called — without it the test binary exits
// immediately and no tests in this package ever execute.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func compareProducts(expected, response_data []repo.Product) string {
	if len(expected) != len(response_data) {
		return fmt.Sprintf("length: expected %d, response_data %d", len(expected), len(response_data))
	}
	var sb strings.Builder
	for i := range response_data {
		e, r := expected[i], response_data[i]
		if e.ID != r.ID || e.Name != r.Name || e.Quantity != r.Quantity || e.Price != r.Price {
			fmt.Fprintf(&sb, "[%d] expected {ID:%d Name:%q Price:%d Quantity:%d} response_data {ID:%d Name:%q Price:%d Quantity:%d}\n",
				i, e.ID, e.Name, e.Price, e.Quantity, r.ID, r.Name, r.Price, r.Quantity)
		}
	}
	return sb.String()
}

func makeRequest(t *testing.T, srv *httptest.Server) []repo.Product {
	t.Helper()
	res, err := http.Get(srv.URL + "/products")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	// products will be filled by Unmarshal from the JSON response body (not the DB directly)
	var products []repo.Product
	if err := json.Unmarshal(body, &products); err != nil {
		t.Fatalf("decode response body %q: %v", body, err)
	}
	return products
}

func TestListProducts_EmptyWhenThereAreNoProducts(t *testing.T) {
	pool := testdb.New(t, "../adapters/postgresql/migrations")

	// Make request and get the response
	router := chi.NewRouter()
	router.Get("/products", NewHandler(NewService(repo.New(pool))).ListProducts)
	srv := httptest.NewServer(router)
	defer srv.Close()

	response_products := makeRequest(t, srv)
	if len(response_products) != 0 {
		t.Errorf("expected no products, recieved %d: %v", len(response_products), response_products)
	}
}

func TestListProduct_ReturnsProducts(t *testing.T) {

	// create products
	pool := testdb.New(t, "../adapters/postgresql/migrations")
	factory := testdb.NewFactory(t, pool)
	expected_products := factory.Products(3)

	router := chi.NewRouter()
	router.Get("/products", NewHandler(NewService(repo.New(pool))).ListProducts)
	srv := httptest.NewServer(router)
	defer srv.Close()

	response_products := makeRequest(t, srv)

	// Compare only fields that survive the JSON round-trip
	// (avoids time.Time DeepEqual flakiness on created_at)
	if diff := compareProducts(expected_products, response_products); diff != "" {
		t.Errorf("products mismatch (-want +got):\n%s", diff)
	}
}

func TestProductLists_ChecksOrderOfProduct(t *testing.T) {
	pool := testdb.New(t, "../adapters/postgresql/migrations")
	factory := testdb.NewFactory(t, pool)
	expected_products := factory.Products(3)

	router := chi.NewRouter()
	router.Get("/products", NewHandler(NewService(repo.New(pool))).ListProducts)
	srv := httptest.NewServer(router)
	defer srv.Close()

	response_products := makeRequest(t, srv)

	expected_ids := make([]int64, len(expected_products))
	for i, p := range expected_products {
		expected_ids[i] = p.ID
	}

	response_products_ids := make([]int64, len(expected_products))
	for i, p := range response_products {
		response_products_ids[i] = p.ID
	}

	for i := range response_products_ids {
		if response_products[i].ID != expected_ids[i] {
			t.Errorf("Expected products with ID order of %v but recieved %v", expected_ids, response_products_ids)
		}
	}
}

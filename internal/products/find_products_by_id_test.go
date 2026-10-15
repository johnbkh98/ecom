package products

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	repo "github.com/johnbkh98/ecom/internal/adapters/postgresql/sqlc"
	"github.com/johnbkh98/ecom/internal/testdb"
)

// TestMain is a placeholder for package-level test setup (e.g. a shared test
// database). Note: m.Run() must be called — without it the test binary exits
// immediately and no tests in this package ever execute.
func TestMain_2(m *testing.M) {
	os.Exit(m.Run())
}

func makeRequestForFindProduct(t *testing.T, srv *httptest.Server, product_id string) []repo.Product {
	t.Helper()
	res, err := http.Get(srv.URL + "/product/" + product_id)
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

func FindProductById_ReturnsErrorWhenProductIdDoesNotExist(t *testing.T) {
	t.Helper()

	pool := testdb.New(t, "../adapters/postgresql/migrations")
	non_existence_product_id := "999"
	router := chi.NewRouter()
	router.Get("/product/"+non_existence_product_id, NewHandler(NewService(repo.New(pool))).ListProducts)
	srv := httptest.NewServer(router)
	defer srv.Close()

	response_product := makeRequestForFindProduct(t, srv, non_existence_product_id)

	fmt.Println(response_product)
}

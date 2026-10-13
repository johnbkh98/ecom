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
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

/**
---------------------------
TO DO:
---------------------------
Get rid of all current tests.
Add the following test cases:
- TestListProducts_ReturnsEmptyWhenThereAreNoProducts
	-- this should hit the /products endpoint
	-- get the response from the endpoint
	-- assert empty
- TestListProducts_ReturnListOfProducts
	-- Creat lets say 3 products
	-- hit the /products endpoint
	-- get the response from the endpoint
	-- assert response contains the 3 products
*/

func TestListProducts_EmptyWhenThereAreNoProducts(t *testing.T) {
	pool := testdb.New(t, "../adapters/postgresql/migrations")

	// Make request and get the response
	router := chi.NewRouter()
	router.Get("/products", NewHandler(NewService(repo.New(pool))).ListProducts)

	srv := httptest.NewServer(router)
	defer srv.Close()

	response, err := http.Get(srv.URL + "/products")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	// assert Response
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status code = %d, recieved %d", http.StatusOK, response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	var products []repo.Product
	if err := json.Unmarshal(body, &products); err != nil {
		t.Fatalf("decode response body %q: %v", body, err)
	}
	if len(products) != 0 {
		t.Errorf("expected no products, got %d: %s", len(products), body)
	}
	// make a reuseable function makeRequest?
}

func TestListProduct_ReturnsProducts(t *testing.T) {

	pool := testdb.New(t, "../adapters/postgresql/migrations")
	factory := testdb.NewFactory(t, pool)
	products := factory.Product(3)
	fmt.Println(products)
	// Make request and get the response
	router := chi.NewRouter()
	router.Get("/products", NewHandler(NewService(repo.New(pool))).ListProducts)

	srv := httptest.NewServer(router)
	defer srv.Close()

	response, err := http.Get(srv.URL + "/products")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(response)
}

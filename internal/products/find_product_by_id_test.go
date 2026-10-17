package products

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	repo "github.com/johnbkh98/ecom/internal/adapters/postgresql/sqlc"
	"github.com/johnbkh98/ecom/internal/testdb"
)

type responseResults struct {
	Status  int
	Product *repo.Product
	Message string
}

func makeRequestForFindProduct(t *testing.T, srv *httptest.Server, product_id string) responseResults {
	t.Helper()
	// make request and store the results in the responseResults for easier assertions in tests.
	res, err := http.Get(srv.URL + "/product/" + product_id)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	results := responseResults{Status: res.StatusCode, Message: string(body)}
	if res.StatusCode == http.StatusOK {
		var p repo.Product
		if err := json.Unmarshal(body, &p); err != nil {
			t.Fatalf("decode response body %q: %v", body, err)
		}
		results.Product = &p
	}
	return results
}

func setupTestServerForFindProductById(t *testing.T) (*httptest.Server, *pgxpool.Pool) {
	pool := testdb.New(t, "../adapters/postgresql/migrations")
	router := chi.NewRouter()
	router.Get("/product/{product_id}", NewHandler(NewService(repo.New(pool))).FindProductById)
	srv := httptest.NewServer(router)
	return srv, pool
}

func TestFindProductById_ReturnsErrorWhenProductIdDoesNotExist(t *testing.T) {
	srv, _ := setupTestServerForFindProductById(t)
	defer srv.Close()

	response := makeRequestForFindProduct(t, srv, "999")
	if response.Status != http.StatusNotFound {
		t.Errorf("Expected status code 404 but received: %v", response.Status)
	}
	if response.Product != nil {
		t.Errorf("Expected not Product instead received: %v", response.Product)
	}
	message := "Unable to find product."
	if strings.TrimSpace(response.Message) != message {
		t.Errorf("Expected message: %v received: %v", message, response.Message)
	}
}

func TestFindProductById_ReturnsErrorWhenProductIdCanNotBeProcessed(t *testing.T) {
	srv, _ := setupTestServerForFindProductById(t)
	defer srv.Close()

	response := makeRequestForFindProduct(t, srv, "invalid_product")
	if response.Status != http.StatusBadRequest {
		t.Errorf("Expected status code %v but received: %v", http.StatusBadRequest, response.Status)
	}
	if response.Product != nil {
		t.Errorf("Expected not Product instead received: %v", response.Product)
	}
	message := "Unable to process input."
	if strings.TrimSpace(response.Message) != message {
		t.Errorf("Expected message: %v received: %v", message, response.Message)
	}
}

func TestFindProductById_ReturnsProduct(t *testing.T) {
	srv, pool := setupTestServerForFindProductById(t)
	factory := testdb.NewFactory(t, pool)
	expected_product := factory.Product()
	defer srv.Close()

	response := makeRequestForFindProduct(t, srv, strconv.FormatInt(expected_product.ID, 10))
	if response.Status != http.StatusOK {
		t.Errorf("Expected status code %v. received %v.", http.StatusOK, response.Status)
	}
	product := response.Product
	if product == nil {
		t.Fatal("No product was returned from response.")
	}
	if product.ID != expected_product.ID || product.Name != expected_product.Name || product.Price != expected_product.Price || product.Quantity != expected_product.Quantity {
		t.Errorf("Expected product: %v. received: %v", expected_product, product)
	}
}

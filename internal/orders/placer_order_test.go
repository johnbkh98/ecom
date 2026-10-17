package orders

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	repo "github.com/johnbkh98/ecom/internal/adapters/postgresql/sqlc"
	"github.com/johnbkh98/ecom/internal/testdb"
)

type responseResults struct {
	Status int
	Body   string
}

func makeServerForPlaceOrder(t *testing.T) (*httptest.Server, *pgxpool.Pool) {
	pool := testdb.New(t, "../adapters/postgresql/migrations")
	router := chi.NewRouter()
	router.Post("/order", NewHandler(NewService(repo.New(pool), pool)).PlaceOrder)
	srv := httptest.NewServer(router)

	return srv, pool
}

func makeRequestForPlaceOrder(t *testing.T, srv *httptest.Server, payload []byte, message *string) responseResults {
	t.Helper()
	res, err := http.Post(srv.URL+"/order", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	*message = string(body)
	results := responseResults{Status: res.StatusCode, Body: *message}
	return results
}

/*
	Test cases
	- When customer id is not provided
	- when customer id is invalid
	- when items are not provided
	- when items is empty
	- when items are invalid
	- when product does not exist
*/

func TestPlaceOrder_ItReturnsErrorWhenCustomerIdIsNotProvided(t *testing.T) {
	var payload = []byte(`{"items": [{"productId": 1, "quantity": 2}]}`)
	srv, _ := makeServerForPlaceOrder(t)
	message := ""
	response_results := makeRequestForPlaceOrder(t, srv, payload, &message)
	if response_results.Status != http.StatusBadRequest {
		t.Errorf("Expected status code: %v. Recieved: %v.", http.StatusBadRequest, response_results.Status)
	}

	expected_message := "Customer ID is required."
	response_body := strings.TrimSpace(response_results.Body)
	if response_body != expected_message {
		t.Errorf("Expected message: %v Recieved: %v.", expected_message, response_body)
	}
}

// when customer id is invalid

func TestPlaceOrder_ItReturnsErrorWhenItemsAreNotProvided(t *testing.T) {
	// var payload = []byte(`{"CustomerID": 34}`)
}

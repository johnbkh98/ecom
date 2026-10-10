package products

import (
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	repo "github.com/johnbkh98/ecom/internal/adapters/postgresql/sqlc"
	"github.com/johnbkh98/ecom/internal/json"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (handler *handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	// Call service and return products in an http response

	products, err := handler.service.ListProducts(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// A nil slice would be encoded as `null`; clients expect an empty array.
	if products == nil {
		products = []repo.Product{}
	}

	json.Write(w, http.StatusOK, products)
}

func (handler *handler) FindProductById(w http.ResponseWriter, r *http.Request) {
	// call service and return product in an http response
	product_id := chi.URLParam(r, "product_id")
	product_id_int, err_int_conv := strconv.ParseInt(product_id, 10, 64)
	if err_int_conv != nil {
		// Handle string converstion to int64 error
		log.Println(err_int_conv)
		http.Error(w, "Unable to proccess input", http.StatusInternalServerError)
		return
	}
	product, err_find_product := handler.service.FindProductById(r.Context(), product_id_int)
	if err_find_product != nil {
		// Handle error when finding product
		log.Println(err_find_product)
		http.Error(w, "Unable to find product.", http.StatusNotFound)
		return
	}

	json.Write(w, http.StatusOK, product)
}

package products

import (
	"log"
	"net/http"

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

	json.Write(w, http.StatusOK, products)
}

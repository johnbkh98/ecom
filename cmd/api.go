package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	repo "github.com/johnbkh98/ecom/internal/adapters/postgresql/sqlc"
	"github.com/johnbkh98/ecom/internal/orders"
	"github.com/johnbkh98/ecom/internal/products"
)

// mount endpoints
func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)              // important for rate limiting
	r.Use(middleware.RealIP)                 // important for rate limiting analytics - deprecated
	r.Use(middleware.ClientIPFromRemoteAddr) // pick one ClientIPFrom* based on your infra, see below
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer) // recover from crashes

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	// Endpoints
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("All Good"))
	})

	productService := products.NewService(repo.New(app.db))
	productHandler := products.NewHandler(productService)
	orderService := orders.NewService(repo.New(app.db), app.db)
	ordersHandler := orders.NewHandler(orderService)

	r.Get("/products", productHandler.ListProducts)
	r.Get("/product/{product_id}", productHandler.FindProductById)
	r.Post("/order", ordersHandler.PlaceOrder)

	return r
}

// run servers
func (app *application) run(handler http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      handler,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	log.Printf("Server has started at %s", app.config.addr)
	return srv.ListenAndServe()
}

type application struct {
	config config
	// logger
	db *pgxpool.Pool
}

type config struct {
	addr string
	db   dbConfig
}

type dbConfig struct {
	dsn string
}

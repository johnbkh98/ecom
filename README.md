Run

`docker compose up` to start db container
`go run cmd/*.go` to start application
`docker compose down` to stop db container
`goose up` to run migration
`goose down` to rollback migration
`goose -s create action_table_name sql` to create a migration eg. `goose -s create create__order_table sql`
`sqlc generate` to generate sql queries to go

`docker exec ecom-postgres psql -U postgres -c 'CREATE DATABASE ecom_test'` - to create test db

### TO DO (features):
- [x] Files rename. e.g product_handlers to avoid tab ambiguity
- [ ] update product quantity after an order has been placed
- [ ] Add tests
- [ ] Create a setup script that runs neccessary commands that get the project up and running
- [ ] Create an endpoint to create products - `r.Create("/product")`
- [ ] Create a new endpoint to get an order - `r.Get("/order/{order_id}")`
- [ ] Create a new endpoint to get all orders - `r.Get("/orders")`
- [ ] Create a customers table
- [ ] Create an endpoint to get all customers - `r.Get("/cutomers")`
- [ ] Create an endpoint to get all orders from a customer - `r.Get("/orders/cutomers/{customer_id}")`

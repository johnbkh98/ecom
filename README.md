Run

`go run cmd/*.go` to start application
`docker compose up` to start db container
`docker compose down` to stop db container
`goose up` to run migration
`goose down` to rollback migration
`goose -s create action_table_name sql` to create a migration eg. `goose -s create create__order_table sql`
`sqlc generate` to generate sql queries to go
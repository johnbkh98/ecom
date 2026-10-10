-- name: ListProducts :many

SELECT
    *
FROM
    products;

-- name: FindProductById :one
SELECT * FROM products WHERE id = $1;

-- name: CreateOrder :one
INSERT INTO orders (
  customer_id
) VALUES ($1) RETURNING *;

-- name: CreateOrderItem :one
INSERT INTO order_items (order_id, product_id, quantity, price)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: CreateProduct :one
INSERT INTO products (name, quantity, created_at, price)
VALUES ($1, $2, $3, $4) RETURNING *;
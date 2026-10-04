-- name: ListProducts :many

SELECT
    *
FROM
    products;

-- name: FindProductById :one
SELECT * FROM products WHERE id = $1;
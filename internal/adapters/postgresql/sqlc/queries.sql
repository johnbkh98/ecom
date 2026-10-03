-- name: ListProducts :many

SELECT
    *
FROM
    products;

-- name: FindProudctByID :one
SELECT * FROM products WHERE id = $1;
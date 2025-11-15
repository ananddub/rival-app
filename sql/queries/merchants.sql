-- name: CreateMerchant :one
INSERT INTO merchants (
    name, email, phone, category, discount_percentage, is_active
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetMerchantByID :one
SELECT * FROM merchants WHERE id = $1;

-- name: GetMerchantByEmail :one
SELECT * FROM merchants WHERE email = $1;

-- name: UpdateMerchant :exec
UPDATE merchants SET
    name = $2,
    phone = $3,
    category = $4,
    discount_percentage = $5,
    is_active = $6,
    updated_at = NOW()
WHERE id = $1;

-- name: ListActiveMerchants :many
SELECT * FROM merchants WHERE is_active = true ORDER BY name;

-- name: GetMerchantsByCategory :many
SELECT * FROM merchants WHERE category = $1 AND is_active = true ORDER BY name;

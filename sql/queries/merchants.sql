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

-- name: GetMerchantTransactions :many
SELECT * FROM transactions 
WHERE merchant_id = $1 
ORDER BY created_at DESC 
LIMIT $2 OFFSET $3;

-- name: GetMerchantCustomers :many
SELECT DISTINCT u.* FROM users u
JOIN transactions t ON u.id = t.user_id
WHERE t.merchant_id = $1
ORDER BY u.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountMerchants :one
SELECT COUNT(*) FROM merchants;

-- name: CountActiveMerchants :one
SELECT COUNT(*) FROM merchants WHERE is_active = true;

-- name: GetAllMerchants :many
SELECT * FROM merchants 
ORDER BY created_at DESC 
LIMIT $1 OFFSET $2;

-- name: GetMerchantAddresses :many
SELECT * FROM merchant_addresses 
WHERE merchant_id = $1 
ORDER BY created_at DESC;

-- name: CreateMerchantAddress :one
INSERT INTO merchant_addresses (
    merchant_id, street, city, state, postal_code, country, latitude, longitude, is_primary
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: CreateOffer :one
INSERT INTO offers (
    merchant_id, title, description, discount_percentage, min_amount, max_discount, valid_from, valid_until
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetMerchantOffers :many
SELECT * FROM offers 
WHERE merchant_id = $1 AND is_active = true
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetOfferByID :one
SELECT * FROM offers WHERE id = $1;

-- name: UpdateOffer :exec
UPDATE offers SET
    title = $2,
    description = $3,
    discount_percentage = $4,
    min_amount = $5,
    max_discount = $6,
    valid_from = $7,
    valid_until = $8,
    is_active = $9,
    updated_at = NOW()
WHERE id = $1;

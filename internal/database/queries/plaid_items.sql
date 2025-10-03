-- name: CreatePlaidItem :one
INSERT INTO plaid_items (user_id, access_token, item_id, institution_name)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, access_token, item_id, institution_name, transactions_cursor, created_at, updated_at;

-- name: GetPlaidItemsByUserID :many
SELECT id, user_id, access_token, item_id, institution_name, transactions_cursor, created_at, updated_at
FROM plaid_items
WHERE user_id = $1;

-- name: GetPlaidItemByItemID :one
SELECT id, user_id, access_token, item_id, institution_name, transactions_cursor, created_at, updated_at
FROM plaid_items
WHERE item_id = $1;

-- name: UpdatePlaidItemAccessToken :one
UPDATE plaid_items
SET access_token = $2, updated_at = now()
WHERE item_id = $1
RETURNING id, user_id, access_token, item_id, institution_name, transactions_cursor, created_at, updated_at;

-- name: UpdatePlaidItemCursor :one
UPDATE plaid_items
SET transactions_cursor = $2, updated_at = now()
WHERE item_id = $1
RETURNING id, user_id, access_token, item_id, institution_name, transactions_cursor, created_at, updated_at;

-- name: DeletePlaidItem :exec
DELETE FROM plaid_items
WHERE id = $1;

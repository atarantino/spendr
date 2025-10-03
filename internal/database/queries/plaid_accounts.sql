-- name: CreatePlaidAccount :one
INSERT INTO plaid_accounts (plaid_item_id, account_id, name, official_name, type, subtype)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, plaid_item_id, account_id, name, official_name, type, subtype, created_at, updated_at;

-- name: GetPlaidAccountsByItemID :many
SELECT id, plaid_item_id, account_id, name, official_name, type, subtype, created_at, updated_at
FROM plaid_accounts
WHERE plaid_item_id = $1;

-- name: GetPlaidAccountByAccountID :one
SELECT id, plaid_item_id, account_id, name, official_name, type, subtype, created_at, updated_at
FROM plaid_accounts
WHERE account_id = $1;

-- name: CreateTransaction :one
INSERT INTO transactions (
    user_id, plaid_account_id, transaction_id, account_id, amount, date,
    authorized_date, name, merchant_name, pending, payment_channel,
    transaction_code, iso_currency_code, unofficial_currency_code,
    location, payment_meta, personal_finance_category, counterparties
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
ON CONFLICT (transaction_id) DO NOTHING
RETURNING id, user_id, plaid_account_id, transaction_id, account_id, amount, date,
    authorized_date, name, merchant_name, pending, payment_channel,
    transaction_code, iso_currency_code, unofficial_currency_code,
    location, payment_meta, personal_finance_category, counterparties, created_at, updated_at;

-- name: GetTransactionsByUserID :many
SELECT id, user_id, plaid_account_id, transaction_id, account_id, amount, date,
    authorized_date, name, merchant_name, pending, payment_channel,
    transaction_code, iso_currency_code, unofficial_currency_code,
    location, payment_meta, personal_finance_category, counterparties, created_at, updated_at
FROM transactions
WHERE user_id = $1
ORDER BY date DESC;

-- name: GetTransactionByID :one
SELECT id, user_id, plaid_account_id, transaction_id, account_id, amount, date,
    authorized_date, name, merchant_name, pending, payment_channel,
    transaction_code, iso_currency_code, unofficial_currency_code,
    location, payment_meta, personal_finance_category, counterparties, created_at, updated_at
FROM transactions
WHERE id = $1;

-- name: GetTransactionByPlaidTransactionID :one
SELECT id, user_id, plaid_account_id, transaction_id, account_id, amount, date,
    authorized_date, name, merchant_name, pending, payment_channel,
    transaction_code, iso_currency_code, unofficial_currency_code,
    location, payment_meta, personal_finance_category, counterparties, created_at, updated_at
FROM transactions
WHERE transaction_id = $1;

-- name: GetUncategorizedTransactionsByUserID :many
SELECT t.id, t.user_id, t.plaid_account_id, t.transaction_id, t.account_id, t.amount, t.date,
    t.authorized_date, t.name, t.merchant_name, t.pending, t.payment_channel,
    t.transaction_code, t.iso_currency_code, t.unofficial_currency_code,
    t.location, t.payment_meta, t.personal_finance_category, t.counterparties, t.created_at, t.updated_at
FROM transactions t
LEFT JOIN transaction_categorizations tc ON t.id = tc.transaction_id AND tc.wallet_id = $2
WHERE t.user_id = $1 AND tc.id IS NULL
ORDER BY t.date DESC;

-- name: GetNextUncategorizedTransactionByUserID :one
SELECT t.id, t.user_id, t.plaid_account_id, t.transaction_id, t.account_id, t.amount, t.date,
    t.authorized_date, t.name, t.merchant_name, t.pending, t.payment_channel,
    t.transaction_code, t.iso_currency_code, t.unofficial_currency_code,
    t.location, t.payment_meta, t.personal_finance_category, t.counterparties, t.created_at, t.updated_at
FROM transactions t
LEFT JOIN transaction_categorizations tc ON t.id = tc.transaction_id AND tc.wallet_id = $2
WHERE t.user_id = $1 AND tc.id IS NULL
ORDER BY t.date DESC, t.id DESC
LIMIT 1;

-- name: UpdateTransactionPendingStatus :exec
UPDATE transactions
SET pending = $2, updated_at = now()
WHERE transaction_id = $1;

-- name: GetTransactionsByUserIDPaginated :many
SELECT id, user_id, plaid_account_id, transaction_id, account_id, amount, date,
    authorized_date, name, merchant_name, pending, payment_channel,
    transaction_code, iso_currency_code, unofficial_currency_code,
    location, payment_meta, personal_finance_category, counterparties, created_at, updated_at
FROM transactions
WHERE user_id = $1
ORDER BY date DESC
LIMIT $2 OFFSET $3;

-- name: CountTransactionsByUserID :one
SELECT COUNT(*) FROM transactions WHERE user_id = $1;

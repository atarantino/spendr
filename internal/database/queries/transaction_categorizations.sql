-- name: CreateTransactionCategorization :one
INSERT INTO transaction_categorizations (transaction_id, wallet_id, category_type, categorized_by_user_id)
VALUES ($1, $2, $3, $4)
RETURNING id, transaction_id, wallet_id, category_type, categorized_by_user_id, categorized_at;

-- name: GetCategorizationByTransactionAndWallet :one
SELECT id, transaction_id, wallet_id, category_type, categorized_by_user_id, categorized_at
FROM transaction_categorizations
WHERE transaction_id = $1 AND wallet_id = $2;

-- name: GetSharedTransactionsByWalletID :many
SELECT t.id, t.user_id, t.plaid_account_id, t.transaction_id, t.account_id, t.amount, t.date,
    t.authorized_date, t.name, t.merchant_name, t.pending, t.payment_channel,
    t.transaction_code, t.iso_currency_code, t.unofficial_currency_code,
    t.location, t.payment_meta, t.personal_finance_category, t.counterparties, t.created_at, t.updated_at,
    tc.category_type, tc.categorized_by_user_id, tc.categorized_at
FROM transactions t
JOIN transaction_categorizations tc ON t.id = tc.transaction_id
WHERE tc.wallet_id = $1 AND tc.category_type = 'shared'
ORDER BY t.date DESC;

-- name: DeleteTransactionCategorization :exec
DELETE FROM transaction_categorizations
WHERE transaction_id = $1 AND wallet_id = $2;

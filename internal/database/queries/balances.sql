-- name: UpsertBalance :one
INSERT INTO balances (wallet_id, user_id, net_balance, last_updated_at)
VALUES ($1, $2, $3, now())
ON CONFLICT (wallet_id, user_id)
DO UPDATE SET net_balance = $3, last_updated_at = now()
RETURNING wallet_id, user_id, net_balance, last_updated_at;

-- name: GetBalanceByWalletAndUser :one
SELECT wallet_id, user_id, net_balance, last_updated_at
FROM balances
WHERE wallet_id = $1 AND user_id = $2;

-- name: GetBalancesByWalletID :many
SELECT b.wallet_id, b.user_id, b.net_balance, b.last_updated_at, u.name, u.email
FROM balances b
JOIN users u ON b.user_id = u.id
WHERE b.wallet_id = $1;

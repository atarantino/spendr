-- name: CreateWallet :one
INSERT INTO wallets (name)
VALUES ($1)
RETURNING id, name, created_at, updated_at;

-- name: GetWalletByID :one
SELECT id, name, created_at, updated_at
FROM wallets
WHERE id = $1;

-- name: AddWalletMember :exec
INSERT INTO wallet_members (wallet_id, user_id)
VALUES ($1, $2);

-- name: GetWalletMembersByWalletID :many
SELECT wm.wallet_id, wm.user_id, wm.joined_at, u.name, u.email
FROM wallet_members wm
JOIN users u ON wm.user_id = u.id
WHERE wm.wallet_id = $1;

-- name: GetWalletByUserID :one
SELECT w.id, w.name, w.created_at, w.updated_at
FROM wallets w
JOIN wallet_members wm ON w.id = wm.wallet_id
WHERE wm.user_id = $1
LIMIT 1;

-- name: RemoveWalletMember :exec
DELETE FROM wallet_members
WHERE wallet_id = $1 AND user_id = $2;

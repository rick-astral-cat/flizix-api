-- name: CreateAccount :one
INSERT INTO accounts (
    name,
    type,
    user_id
) VALUES (?, ?, ?) RETURNING *;

-- name: GetAccountByID :one
SELECT * FROM accounts
WHERE id = ? AND user_id = ? AND deleted_at IS NULL
LIMIT 1;

-- name: ListAccountsByUser :many
SELECT * FROM accounts
WHERE user_id = ? AND deleted_at IS NULL
ORDER BY name ASC;

-- name: SoftDeleteAccount :exec
UPDATE accounts SET deleted_at = CURRENT_TIMESTAMP
WHERE id = ? AND user_id = ?;
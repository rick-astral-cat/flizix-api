-- name: CreateAccountType :one
INSERT INTO account_types (
    name,
    user_id,
    is_system
) VALUES (?, ?, ?) RETURNING *;

-- name: ListAccountTypesByUser :many
SELECT * FROM account_types
WHERE (user_id IS NULL OR user_id = ?) AND deleted_at IS NULL
ORDER BY is_system DESC, name ASC;

-- name: ListSystemAccountTypes :many
SELECT * FROM  account_types
WHERE user_id IS NULL AND deleted_at IS NULL
ORDER BY name ASC;

-- name: GetAccountTypeByUser :one
SELECT * FROM account_types
WHERE id = ? AND (user_id IS NULL OR user_id = ?) AND deleted_at IS NULL;

-- name: SoftDeleteAccountTypeByUser :exec
UPDATE account_types SET deleted_at = CURRENT_TIMESTAMP
WHERE id = ? AND user_id = ? AND is_system = 0;
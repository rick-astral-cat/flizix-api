-- name: GetUserById :one
-- Used to get user profile once registered
SELECT * FROM users WHERE id = ? AND deleted_at IS NULL LIMIT 1;

-- name: GetUserByTelegramId :one
SELECT * FROM users WHERE telegram_id = ? AND deleted_at IS NULL LIMIT 1;

-- name: GetUserByPassKey :one
SELECT * FROM users WHERE passkey_id = ? AND deleted_at IS NULL LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ? AND deleted_at IS NULL LIMIT 1;

-- name: CreateUserWithTelegram :one
-- Initial user registration
INSERT INTO users (
    name,
    email,
    telegram_id
) VALUES (?, ?, ?) RETURNING *;

-- name: CreateUserWithPasskey :one
INSERT INTO users (
    name,
    email,
    passkey_id
) VALUES (?, ?, ?) RETURNING *;

-- name: UpdateUserPasskey :exec
UPDATE users SET passkey_id = ? WHERE id = ? AND deleted_at IS NULL;

-- name: SoftDeleteUser :exec
UPDATE users SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?;

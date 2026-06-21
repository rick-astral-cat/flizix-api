-- name: CreateCard :one
INSERT INTO cards (
    name,
    type,
    credit_limit,
    cutoff_date,
    account_id,
    user_id
) VALUES (?,?,?,?,?,?)
RETURNING *;

-- name: GetCardByID :one
SELECT * FROM cards
WHERE id = ? AND user_id = ? AND deleted_at IS NULL
LIMIT 1;

-- name: ListCardsByUser :many
SELECT * FROM cards
WHERE user_id = ? AND deleted_at IS NULL
ORDER BY name ASC;

-- name: UpdateCard :one
UPDATE cards
SET name = ?,
    type = ?,
    credit_limit = ?,
    cutoff_date = ?
WHERE id = ? AND user_id = ? AND deleted_at is NULL
RETURNING *;

-- name: SoftDeleteCard :exec
UPDATE cards SET deleted_at = CURRENT_TIMESTAMP
WHERE id = ? AND user_id = ?;

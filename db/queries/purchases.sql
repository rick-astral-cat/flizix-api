-- name: CreatePurchase :one
INSERT INTO purchases (
    purchase,
    date,
    amount,
    carry_over_next_month,
    comment,
    card_id,
    account_id,
    project_id,
    user_id
) VALUES (?,?,?,?,?,?,?,?,?) RETURNING *;

-- name: GetPurchaseById :one
SELECT * FROM purchases WHERE id = ? AND user_id = ? AND deleted_at IS NULL;

-- name: ListPurchasesByUserId :many
SELECT * FROM purchases WHERE user_id = ? AND deleted_at IS NULL ORDER BY date DESC;

-- name: UpdatePurchase :one
UPDATE purchases SET
purchase = ?, date = ?, amount = ?, amount_paid = ?, carry_over_next_month = ?,
comment = ?, card_id = ?, account_id = ?, project_id = ?
WHERE id = ? AND user_id = ?
RETURNING *;

-- name: SoftDeletePurchase :exec
UPDATE purchases SET deleted_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?;
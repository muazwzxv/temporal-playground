-- name: GetUserByID :one
SELECT * FROM users WHERE id = ?;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CreateUser :execresult
INSERT INTO users (name, description, status, created_at, updated_at)
VALUES (?, ?, ?, NOW(), NOW());

-- name: UpdateUser :exec
UPDATE users
SET name = ?, description = ?, status = ?, updated_at = NOW()
WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: GetUsersByStatus :many
SELECT * FROM users
WHERE status = ?
ORDER BY created_at DESC;

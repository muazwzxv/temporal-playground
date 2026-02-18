-- name: GetUserByUUID :one
SELECT * FROM users WHERE user_uuid = ?;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CreateUser :execresult
INSERT INTO users
    (user_uuid, name, status, created_at, updated_at)
VALUES
    (?, ?, ?, NOW(), NOW());

-- -- name: UpdateUser :exec
-- UPDATE users
-- SET name = ?, description = ?, status = ?, updated_at = NOW()
-- WHERE id = ?;

-- -- name: DeleteUser :exec
-- DELETE FROM users WHERE id = ?;

-- -- name: CountUsers :one
-- SELECT COUNT(*) FROM users;

-- -- name: GetUsersByStatus :many
-- SELECT * FROM users
-- WHERE status = ?
-- ORDER BY created_at DESC;

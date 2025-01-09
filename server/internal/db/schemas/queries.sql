-- name: CreateFile :one
INSERT INTO files (id, filename, expires_at)
VALUES (?, ?, datetime('now', '+' || CAST(sqlc.arg(expiration) as INTEGER) || ' minutes'))
RETURNING *;

-- name: GetFile :one
SELECT *
FROM files
WHERE id = ?
AND expires_at > CURRENT_TIMESTAMP
LIMIT 1;

-- name: GetExpiredFiles :many
SELECT *
FROM files
WHERE expires_at < CURRENT_TIMESTAMP
ORDER BY expires_at;

-- name: DeleteFile :exec
DELETE FROM files
WHERE id = ?;
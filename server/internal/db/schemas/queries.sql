-- name: CreateFile :one
INSERT INTO files (id, filename)
VALUES (?, ?)
RETURNING *;

-- name: GetFile :one
SELECT *
FROM files
WHERE id = ?
LIMIT 1;
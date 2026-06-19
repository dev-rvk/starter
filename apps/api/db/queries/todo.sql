-- name: CreateTodo :one
INSERT INTO todo (id, title, completed, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateTodo :one
UPDATE todo
SET title = $2, completed = $3, updated_at = $4
WHERE id = $1
RETURNING *;

-- name: DeleteTodo :exec
DELETE FROM todo
WHERE id = $1;

-- name: GetAllTodos :many
SELECT * FROM todo
ORDER BY created_at DESC;
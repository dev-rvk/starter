-- name: CreateAccount :one
INSERT INTO accounts (id, email, password_hash, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAccountByEmail :one
SELECT * FROM accounts
WHERE email = $1;

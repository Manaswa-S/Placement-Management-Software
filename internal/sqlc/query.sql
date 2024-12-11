-- name: GetAll :many
SELECT * FROM users;

-- name: SignupUser :one
INSERT INTO users (email, password, role) VALUES ($1, $2, $3)
RETURNING *;

-- name: LoginUser :one
SELECT * FROM users WHERE email = $1;

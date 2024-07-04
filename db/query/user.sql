-- name: CreateUser :one
INSERT INTO users (
  username, hashed_password, full_name, email
) VALUES (
  $1, $2, $3, $4 
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: DeleteUserByUsernameLike :exec
-- for testing purpose
DELETE FROM users
WHERE username LIKE '%' || @username::text || '%';

-- name: InsertUser :one
INSERT INTO users (
	user_id,
	username,
	password_hash
	)
VALUES
	(
	$1,
	$2,
	$3
	)
RETURNING created_at;

-- name: GetUserByUsername :one 
SELECT user_id,password_hash 
FROM users
WHERE username = $1;
-- name: CreateUser :one
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
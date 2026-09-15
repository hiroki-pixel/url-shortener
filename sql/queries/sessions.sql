-- name: InsertSession :exec
INSERT INTO sessions (
    session_id,
    user_id
)
VALUES
(
    $1,
    $2
);

-- name: GetSessionByID :one 
SELECT user_id,expires_at
FROM sessions
WHERE session_id = $1
;

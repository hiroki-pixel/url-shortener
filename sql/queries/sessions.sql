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

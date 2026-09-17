-- name: InsertURL :exec
INSERT INTO urls (
    url_id,
    user_id,
    original_url,
    memo,
    short_code
)
VALUES(
    $1,
    $2,
    $3,
    $4,
    $5
);
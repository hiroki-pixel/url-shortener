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

-- name: ListsURLByUserID :many
SELECT short_code,original_url,memo
FROM urls
WHERE user_id = $1
ORDER BY created_at DESC;
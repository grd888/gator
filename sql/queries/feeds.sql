-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id, last_fetched_at) 
VALUES ($1, $2, $3, $4, $5, $6, NULL) RETURNING *;

-- name: ListFeeds :many
SELECT * FROM feeds;

-- name: GetFeedByUrl :one
SELECT * FROM feeds WHERE url = $1;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY last_fetched_at NULLS FIRST, id
LIMIT 1;

-- name: MarkFeedFetched :one
UPDATE feeds
SET last_fetched_at = $1, updated_at = $1
WHERE id = $2
RETURNING *;

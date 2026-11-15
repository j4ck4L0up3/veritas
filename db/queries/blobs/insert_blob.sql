-- name: CreateBlob :exec
INSERT INTO blobs (
  digest, size, created_at
) VALUES (?, ?, ?);

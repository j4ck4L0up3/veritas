-- name: CreateUploadSession :exec
INSERT INTO upload_sessions (
  uuid, repo, created_at, expires_at
)
VALUES (?, ?, ?, ?);

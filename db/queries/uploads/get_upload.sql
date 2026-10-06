-- name: GetUploadSession :one
SELECT * FROM upload_sessions
WHERE uuid = ? AND repo = ?;

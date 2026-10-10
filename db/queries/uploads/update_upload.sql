-- name: UpdateUploadSessionBytesReceived :exec
UPDATE upload_sessions
  SET bytes_received = ?
WHERE uuid = ? AND repo = ?;

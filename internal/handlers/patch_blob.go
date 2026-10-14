package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	"github.com/j4ck4L0up3/veritas/db"
	"github.com/j4ck4L0up3/veritas/internal/config"
	"github.com/j4ck4L0up3/veritas/internal/helpers"
	"github.com/j4ck4L0up3/veritas/internal/logger"
)

func PatchBlobStreamHandler(lgr logger.Logger, qry *db.Queries, cfg *config.Config) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			lgr.Info("Patching Blob Stream...")

			rawId := r.PathValue("id")

			id, err := uuid.Parse(rawId)
			if err != nil {
				detail := "could not parse uuid from client request"

				lgr.Error(detail, "uuid", rawId, "error", err)

				respErr := helpers.NewError(
					helpers.BLOB_UPLOAD_UNKNOWN,
					helpers.Messages[helpers.BLOB_UPLOAD_UNKNOWN],
					detail,
				)

				body, err := respErr.GetJSON()
				if err != nil {
					lgr.Warnf("could not marshal error json and attach body to request: %v", err)
					w.WriteHeader(http.StatusNotFound)
					return
				}

				w.WriteHeader(http.StatusNotFound)
				if _, err = w.Write(body); err != nil {
					lgr.Warnf("failed to write response error: %v", err)
				}
			}

			repo := r.PathValue("repo")

			session, err := qry.GetUploadSession(r.Context(), id, repo)
			if err != nil {
				detail := "error when attempting to retrieve upload session"

				lgr.Error(detail, "error", err)

				respErr := helpers.NewError(
					helpers.BLOB_UPLOAD_UNKNOWN,
					helpers.Messages[helpers.BLOB_UPLOAD_UNKNOWN],
					detail,
				)

				body, err := respErr.GetJSON()
				if err != nil {
					lgr.Errorf("could not marshal error json and attach body to request: %v", err)
					w.WriteHeader(http.StatusNotFound)
					return
				}

				w.WriteHeader(http.StatusNotFound)
				if _, err := w.Write(body); err != nil {
					lgr.Errorf("could not write error response: %v", err)
				}
				return
			}

			contentRange := r.Header.Get("Content-Range")
			if contentRange != "" {
				offset, err := strconv.Atoi(contentRange)
				if err != nil {
					lgr.Debugf("Content-Range from request: %s", contentRange)
					w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
					return
				}

				if uint64(offset) != session.BytesReceived {
					lgr.Debug(
						"offset received does not equal stored bytes_received",
						"offset",
						offset,
						"bytes_received",
						session.BytesReceived,
					)
					w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
					return
				}
			}

			filename := filepath.Join(cfg.Locations.UploadPath, id.String())
			f, err := os.OpenFile(
				filename,
				os.O_APPEND|io.SeekEnd,
				0o600,
			)
			if err != nil {
				lgr.Errorf("unable to open upload file: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			defer func() {
				if err := f.Close(); err != nil {
					lgr.Warnf("closing upload file failed: %v", err)
				}
			}()

			written, err := io.Copy(f, r.Body)
			if err != nil {
				lgr.Errorf("unable to copy bytes from request body to upload file: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if err = f.Sync(); err != nil {
				lgr.Warn(
					"failed to sync upload file to disk; may crash on system off/crash",
					"error",
					err,
				)
			}

			params := db.UpdateUploadSessionBytesReceivedParams{
				Uuid:          id,
				Repo:          repo,
				BytesReceived: session.BytesReceived + uint64(written),
			}

			err = qry.UpdateUploadSessionBytesReceived(r.Context(), params)
			if err != nil {
				lgr.Errorf("unable to update bytes received: %v", err)

				err := f.Truncate(int64(session.BytesReceived))
				if err != nil {
					lgr.Errorf(
						"unable to truncate file to the bytes successfully received: %v",
						err,
					)
				}

				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			range_offset := session.BytesReceived + uint64(written) - 1
			w.Header().Add("Location", "/v2/blobs/uploads/"+id.String())
			w.Header().Add("Range", fmt.Sprintf("0-%d", range_offset))
			w.Header().Add("Docker-Upload-UUID", id.String())

			w.WriteHeader(http.StatusOK)
		},
	)
}

package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"uuid"

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
					lgr.Warnf("failed to write error response: %v", err)
				}
				return
			}

			repo := r.PathValue("repo")
			session, err := qry.GetUploadSession(r.Context(), id, repo)
			if err != nil {
				detail := "could not find upload session from id and repo"
				lgr.Error(detail, "uuid", id, "repo", repo, "error", err)

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
				valid := isValidContentRange(contentRange, session.BytesReceived, lgr)
				if !valid {
					w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
					return
				}
			}

			uploadPath := filepath.Join(cfg.Locations.UploadPath, id.String())
			written, f, ok := appendUploadBytes(uploadPath, r.Body, lgr)
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			defer func() {
				if err := f.Close(); err != nil {
					lgr.Warnf("closing upload file failed: %v", err)
				}
			}()

			params := db.UpdateUploadSessionBytesReceivedParams{
				Uuid:          id,
				Repo:          repo,
				BytesReceived: session.BytesReceived + written,
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

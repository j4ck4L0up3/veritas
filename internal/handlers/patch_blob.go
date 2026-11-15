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
			params := appendUploadBytesParams{
				uploadPath:    uploadPath,
				id:            id,
				repo:          repo,
				bytesReceived: session.BytesReceived,
				body:          r.Body,
				ctx:           r.Context(),
				lgr:           lgr,
				qry:           qry,
			}

			rangeOffset, ok := appendUploadBytes(params)
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Add("Location", "/v2/blobs/uploads/"+id.String())
			w.Header().Add("Range", fmt.Sprintf("0-%d", rangeOffset))
			w.Header().Add("Docker-Upload-UUID", id.String())

			w.WriteHeader(http.StatusOK)
		},
	)
}

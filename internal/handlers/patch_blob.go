package handlers

import (
	"fmt"
	"net/http"
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
					helpers.BLOB_UNKNOWN,
					helpers.Messages[helpers.BLOB_UNKNOWN],
					detail,
				)

				body, err := respErr.GetJSON()
				if err != nil {
					lgr.Errorf("could not marshal error json and attach body to request: %v", err)
					w.WriteHeader(http.StatusNotFound)
					return
				}

				w.WriteHeader(http.StatusNotFound)
				w.Write(body)
			}

			repo := r.PathValue("repo")

			session, err := qry.GetUploadSession(r.Context(), id, repo)
			if err != nil {
				detail := "error when attempting to retrieve upload session"

				lgr.Error(detail, "error", err)

				respErr := helpers.NewError(
					helpers.BLOB_UNKNOWN,
					helpers.Messages[helpers.BLOB_UNKNOWN],
					detail,
				)

				body, err := respErr.GetJSON()
				if err != nil {
					lgr.Errorf("could not marshal error json and attach body to request: %v", err)
					w.WriteHeader(http.StatusNotFound)
					return
				}

				w.WriteHeader(http.StatusNotFound)
				w.Write(body)
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

				// start copying from offset, then update bytes_received
			}
		},
	)
}

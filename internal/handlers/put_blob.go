package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"uuid"

	"github.com/j4ck4L0up3/veritas/db"
	"github.com/j4ck4L0up3/veritas/internal/config"
	"github.com/j4ck4L0up3/veritas/internal/helpers"
	"github.com/j4ck4L0up3/veritas/internal/logger"
)

func PutBlobStreamHandler(lgr logger.Logger, qry *db.Queries, cfg *config.Config) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			params := r.URL.Query()
			digest := params.Get("digest")

			if digest == "" || !strings.Contains(digest, ":") {
				detail := "missing or malformed digest from client"
				lgr.Debug(detail)

				respErr := helpers.NewError(
					helpers.DIGEST_INVALID,
					helpers.Messages[helpers.DIGEST_INVALID],
					detail,
				)

				body, err := respErr.GetJSON()
				if err != nil {
					lgr.Warnf("could not marshal error json and attach body to request: %v", err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				w.WriteHeader(http.StatusBadRequest)
				if _, err = w.Write(body); err != nil {
					lgr.Warnf("failed to write error response: %v", err)
				}
				return
			}

			rawId := r.PathValue("id")
			id, err := uuid.Parse(rawId)
			if err != nil {
				detail := "missing or malformed upload id from client"
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
				lgr.Error(detail, "uuid", id.String(), "repo", repo, "error", err)

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

			uploadPath := filepath.Join(cfg.Locations.UploadPath, id.String())

			if r.Body != nil {
				contentRange := r.Header.Get("Content-Range")
				if contentRange != "" {
					valid := isValidContentRange(contentRange, session.BytesReceived, lgr)
					if !valid {
						w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
						return
					}
				}

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

				_, ok := appendUploadBytes(params)
				if !ok {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}

			err, ok := isValidDigest(uploadPath, digest, lgr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if !ok {
				detail := "unsupported digest from client"
				lgr.Debug(detail)

				respErr := helpers.NewError(
					helpers.DIGEST_INVALID,
					helpers.Messages[helpers.DIGEST_INVALID],
					detail,
				)

				body, err := respErr.GetJSON()
				if err != nil {
					lgr.Warnf("could not marshal error json and attach body to request: %v", err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				w.WriteHeader(http.StatusBadRequest)
				if _, err = w.Write(body); err != nil {
					lgr.Warnf("failed to write error response: %v", err)
				}
				return
			}

			uploadHash := strings.Split(digest, ":")[1]
			blobPath := filepath.Join(
				cfg.Locations.BlobPath,
				supportedDigestHash,
				uploadHash[0:2],
				uploadHash[2:4],
				uploadHash,
			)
			dir := filepath.Dir(blobPath)

			if err := os.MkdirAll(dir, 0o755); err != nil {
				lgr.Errorf("unable to make the directories for blob storage: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			err = os.Rename(uploadPath, blobPath)
			if errors.Is(err, os.ErrNotExist) {
				lgr.Info("client blob already exists", "path", blobPath)

				if err := os.Remove(uploadPath); err != nil {
					lgr.Warn("could not remove upload file", "path", uploadPath, "error", err)
				}

				w.Header().Add("Location", fmt.Sprintf("/v2/%s/blobs/%s", repo, digest))
				w.Header().Add("Docker-Content-Digest", digest)
				w.WriteHeader(http.StatusCreated)
				return
			} else if err != nil {
				lgr.Errorf("unable to move upload to blob path: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			d, err := os.Open(dir)
			if err != nil {
				lgr.Errorf("unable to open directory for blob path: %v")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			defer func() {
				if err := d.Close(); err != nil {
					lgr.Warnf("unable to close directory for blob path: %v", err)
				}
			}()

			if err := d.Sync(); err != nil {
				lgr.Warn(
					"unable to sync directory for blob path; may be lost on crash or power loss",
					"error",
					err,
					"path",
					blobPath,
				)
			}

			w.Header().Add("Location", fmt.Sprintf("/v2/%s/blobs/%s", repo, digest))
			w.Header().Add("Docker-Content-Digest", digest)
			w.WriteHeader(http.StatusCreated)
		},
	)
}

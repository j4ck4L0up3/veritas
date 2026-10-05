package handlers

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/j4ck4L0up3/veritas/db"
	"github.com/j4ck4L0up3/veritas/internal/config"
	"github.com/j4ck4L0up3/veritas/internal/logger"
)

func UploadInitiationHandler(
	lgr logger.Logger,
	qry *db.Queries,
	cfg *config.Config,
) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			lgr.Info("Initiating blob upload...")

			id, err := uuid.NewV7()
			if err != nil {
				lgr.Errorf("Failed to generated UUID: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			repo := r.PathValue("repo")

			filename := filepath.Join(cfg.Locations.UploadPath, id.String())
			if _, err = os.OpenFile(
				filename,
				os.O_CREATE|os.O_EXCL,
				0o600,
			); !errors.Is(
				err,
				os.ErrExist,
			) {
				lgr.Errorf("Failed to create upload file: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			params := db.CreateUploadSessionParams{
				Uuid:      id,
				Repo:      repo,
				CreatedAt: uint64(time.Now().Unix()),
				ExpiresAt: uint64(time.Now().Unix()) + cfg.UploadTTL,
			}
			err = qry.CreateUploadSession(r.Context(), params)
			if err != nil {
				lgr.Errorf("Could not create upload session in db: %v", err)

				if err := os.Remove(filename); err != nil {
					lgr.Errorf("Could not remove failed upload session file %s: %v", filename, err)
				}

				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Add("Location", "/v2/blobs/uploads/"+id.String())
			w.Header().Add("Range", "0-0")
			w.Header().Add("Docker-Upload-UUID", id.String())
			w.Header().Add("Content-Length", "0")
			w.WriteHeader(http.StatusAccepted)
		},
	)
}

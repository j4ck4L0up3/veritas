package handlers

import (
	"github.com/j4ck4L0up3/veritas/internal/logger"
	"net/http"
)

func BaseHandler(lgr logger.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			lgr.Info("Performing OCI Registry Handshake...")
			w.Header().Add("Content-Type", "application/json")
			// not required, but recommended to support legacy clients
			w.Header().Add("Docker-Distribution-API-Version", "registry/2.0")

			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{}`)); err != nil {
				lgr.Errorf("Unable to write response: %v", err)
				return
			}
		},
	)
}

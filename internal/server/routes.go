package server

import (
	"net/http"

	"github.com/j4ck4L0up3/veritas/db"
	"github.com/j4ck4L0up3/veritas/internal/config"
	"github.com/j4ck4L0up3/veritas/internal/handlers"
	"github.com/j4ck4L0up3/veritas/internal/logger"
)

func addRoutes(
	mux *http.ServeMux,
	lgr logger.Logger,
	cfg *config.Config,
	qry *db.Queries,
) {
	mux.Handle("/", http.NotFoundHandler())
	mux.Handle("GET /v2/", handlers.BaseHandler(lgr))
	mux.Handle(
		"POST /v2/{repo}/blobs/uploads",
		handlers.UploadInitiationHandler(lgr, qry, cfg),
	)
}

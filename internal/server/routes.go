package server

import (
	"database/sql"
	"github.com/j4ck4L0up3/veritas/internal/config"
	"github.com/j4ck4L0up3/veritas/internal/handlers"
	"github.com/j4ck4L0up3/veritas/internal/logger"
	"net/http"
)

func addRoutes(
	mux *http.ServeMux,
	lgr logger.Logger,
	cfg *config.Config,
	dbConn *sql.DB,
) {
	mux.Handle("/", http.NotFoundHandler())
	mux.Handle("GET /v2/", handlers.BaseHandler(lgr))
}

package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"charm.land/log/v2"
	"github.com/j4ck4L0up3/veritas/internal/config"
	"github.com/j4ck4L0up3/veritas/internal/logger"
)

func New(host string, port uint, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:    net.JoinHostPort(host, strconv.FormatUint(uint64(port), 10)),
		Handler: handler,
	}
}

func NewServerHandler(
	lgr logger.Logger,
	cfg *config.Config,
	dbConn *sql.DB,
) http.Handler {
	mux := http.NewServeMux()

	addRoutes(mux, lgr, cfg, dbConn)

	var handler http.Handler = mux

	// NOTE: middleware can be added here

	return handler
}

func Start(ctx context.Context, server *http.Server) error {
	errChan := make(chan error, 1)
	go func() {
		log.Info(fmt.Sprintf("Starting veritas server at http://%s...", server.Addr))
		if err := server.ListenAndServe(); !errors.Is(
			err, http.ErrServerClosed,
		) {
			errChan <- err
		}
	}()

	sigCtx, stop := signal.NotifyContext(
		ctx,
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	select {
	case err := <-errChan:
		return err
	case <-sigCtx.Done():
		log.Info("Received termination signal, shutting down...")
	}

	shutDownCtx, cancel := context.WithTimeout(
		ctx,
		5*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutDownCtx); err != nil {
		if closeErr := server.Close(); closeErr != nil {
			return errors.Join(err, closeErr)
		}
		return err
	}

	log.Info("Veritas server exited gracefully.")
	return nil
}

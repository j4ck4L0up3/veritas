package main

import (
	"context"
	// "errors"
	"fmt"
	"net/http"
	"os"

	// "path/filepath"

	"charm.land/log/v2"
	"github.com/j4ck4L0up3/veritas/db"
	"github.com/j4ck4L0up3/veritas/internal/cli"
	"github.com/j4ck4L0up3/veritas/internal/config"
	"github.com/j4ck4L0up3/veritas/internal/logger"
	"github.com/j4ck4L0up3/veritas/internal/server"
	"github.com/spf13/cobra"
)

var cfg *config.Config
var lgr logger.Logger
var logFile *os.File
var rootCmd *cobra.Command
var srvr *http.Server

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	cfg = config.Load()

	if err := os.MkdirAll(cfg.Locations.UploadPath, os.FileMode(0o700)); err != nil {
		log.Fatal(fmt.Sprintf("Error creating uploads directory: %v", err))
	}

	if err := os.MkdirAll(cfg.Locations.BlobPath, os.FileMode(0o700)); err != nil {
		log.Fatal(fmt.Sprintf("Error creating blobs directory: %v", err))
	}

	/* if err := os.MkdirAll(cfg.Locations.LogPath, os.FileMode(0o700)); err != nil {
		log.Fatal(fmt.Sprintf("Error creating log directory: %v", err))
	}

	path := filepath.Join(cfg.Locations.LogPath, "veritas.log")
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if errors.Is(err, os.ErrExist) {
	} else if err != nil {
		log.Fatal(fmt.Sprintf("Error opening log file: %v", err))
	}

	logFile = f */
	lgr = logger.New(os.Stdout, cfg.LogLevel, cfg.LogFormat)

	rootCmd = cli.NewCommand("0.0.1")

	dsn := fmt.Sprintf(
		"file:%s?_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000",
		"data/dev.db", // TODO: set db to prod db loc after testing
	)

	dbConn, err := db.RunMigrations(dsn)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error opening db: %v", err))
	}

	srvHandler := server.NewServerHandler(lgr, cfg, dbConn)
	srvr = server.New(cfg.Host, cfg.Port, srvHandler)
}

func run() error {
	/* defer func() {
		if err := logFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing log file: %v", err)
			os.Exit(1)
		}
	}() */

	ctx := context.Background()

	if err := server.Start(ctx, srvr); err != nil {
		return err
	}

	if err := rootCmd.Execute(); err != nil {
		return err
	}

	return nil
}

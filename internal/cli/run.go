package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"charm.land/log/v2"
	"github.com/j4ck4L0up3/veritas/internal/config"
	"github.com/j4ck4L0up3/veritas/internal/logger"
	"github.com/j4ck4L0up3/veritas/internal/server"
	"github.com/spf13/cobra"
)

func newRunCommand(ctx context.Context) *cobra.Command {
	var configPath string
	var logPath string

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "run the server from CLI, creates default config on first run",
		Long:  "Run the server from CLI, creates default config on first run.\nDefault server address is 127.0.0.1:9001 and default config location is ~/.config/veritas/config.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			var lgr logger.Logger

			cfg := getConfig(configPath)

			if logPath != "" {
				lgr = initLogger(cfg, logPath)
			} else {
				lgr = initLogger(cfg, cfg.Locations.LogPath)
			}

			dbPath := filepath.Join(cfg.Locations.DbPath, "veritas.db")
			qry := getDB(dbPath)

			srvHandler := server.NewServerHandler(lgr, cfg, qry)
			srvr := server.New(cfg.Host, cfg.Port, srvHandler)

			return server.Start(ctx, srvr)
		},
	}

	runCmd.Flags().StringVarP(
		&configPath,
		"config",
		"c",
		"",
		"provide a config file for the registry server instance, default",
	)

	runCmd.Flags().StringVarP(
		&logPath,
		"log-file",
		"l",
		"",
		"provide a log file path for the registry server logs. Option to pass 'stdout' or 'stderr' if desired. Default location: ~/.local/share/veritas/veritas.log",
	)

	return runCmd
}

func initLogger(cfg *config.Config, path string) logger.Logger {
	var lgr logger.Logger

	logFile, err := os.OpenFile(
		path,
		os.O_WRONLY|os.O_CREATE,
		0o200,
	)

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("something went wrong while initializing logger from log file: %s", err)
	}

	lgr = logger.New(logFile, cfg.LogLevel, cfg.LogFormat)

	return lgr
}

package cli

import (
	"context"
	"os"
	"path/filepath"

	"github.com/j4ck4L0up3/veritas/internal/logger"
	"github.com/j4ck4L0up3/veritas/internal/server"
	"github.com/spf13/cobra"
)

func newRunCommand(ctx context.Context) *cobra.Command {
	var configPath string
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "run the server from CLI, creates default config on first run",
		Long:  "Run the server from CLI, creates default config on first run.\nDefault server address is 127.0.0.1:9001 and default config location is ~/.config/veritas/config.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := getConfig(configPath)
			lgr := logger.New(os.Stdout, cfg.LogLevel, cfg.LogFormat)
			dbPath := filepath.Join(cfg.Locations.DbPath, "veritas.db")
			conn := getDB(dbPath)

			srvHandler := server.NewServerHandler(lgr, cfg, conn)
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

	return runCmd
}

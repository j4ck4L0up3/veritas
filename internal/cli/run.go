package cli

import (
	"context"
	"database/sql"
	"os"

	"github.com/j4ck4L0up3/veritas/internal/logger"
	"github.com/j4ck4L0up3/veritas/internal/server"
	"github.com/spf13/cobra"
)

func newRunCommand(ctx context.Context, dbConn *sql.DB) *cobra.Command {
	var configPath string
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "run the server from CLI, creates default config on first run",
		Long:  "Run the server from CLI, creates default config on first run.\nDefault server address is 127.0.0.1:9001 and default config location is ~/.config/veritas/config.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := getConfig(configPath)
			lgr := logger.New(os.Stdout, cfg.LogLevel, cfg.LogFormat)

			srvHandler := server.NewServerHandler(lgr, cfg, dbConn)
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

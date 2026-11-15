package cli

import (
	"context"
	"path/filepath"

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
			cfg := getConfig(configPath)
			lgr := cfg.InitLogger(logPath)

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

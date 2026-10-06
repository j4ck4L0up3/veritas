package cli

import (
	"charm.land/log/v2"
	"context"
	"fmt"
	"os"

	"github.com/j4ck4L0up3/veritas/db"
	"github.com/j4ck4L0up3/veritas/internal/config"
	"github.com/spf13/cobra"
)

/*
main command:
veritas

subcommands:
  run      run the server from CLI, creates default config on first run
	init     initialize service config and user systemd unit,
           default config at ~/.config/veritas/config.yaml, default service at ~/.config/systemd/user/veritas.service
	start    start systemd service
	stop     stop systemd service
	enable   enable systemd service
	disable  disable systemd service
	status   show systemd service status

	global flags:
	--help, -h              show help menu
	--version, -v           show version information

	subcommand flags:
	--config, -c {filepath}  provide a config file for the registry server instance, default
*/

func NewCommand(version string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "veritas",
		Short: "veritas is a light-weight, self-hostable container registry server",
		Long: fmt.Sprintf(
			"veritas is a light-weight, self-hostable container registry server\nVersion: %s",
			version,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := cmd.Flags().GetBool("version")
			if err != nil {
				return err
			}

			if v {
				fmt.Printf("Version: %s\n", version)
			}

			return nil
		},
	}

	rootCmd.PersistentFlags().BoolP("version", "v", false, "show version information")

	return rootCmd
}

func SetupCommands(
	root *cobra.Command,
	ctx context.Context,
) {
	root.AddCommand(newRunCommand(ctx))
}

func getConfig(configPath string) *config.Config {
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error retrieving config: %v", err))
	}

	// setup paths
	if err := os.MkdirAll(cfg.Locations.UploadPath, os.FileMode(0o700)); err != nil {
		log.Fatal(fmt.Sprintf("Error creating uploads directory: %v", err))
	}

	if err := os.MkdirAll(cfg.Locations.BlobPath, os.FileMode(0o700)); err != nil {
		log.Fatal(fmt.Sprintf("Error creating blobs directory: %v", err))
	}

	return cfg
}

func getDB(dbPath string) *db.Queries {
	dsn := fmt.Sprintf(
		"file:%s?_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000",
		dbPath,
	)

	conn, err := db.RunMigrations(dsn)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error opening db: %v", err))
	}

	qry := db.New(conn)

	return qry
}

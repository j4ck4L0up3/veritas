package cli

import (
	"fmt"
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
	--config {filepath}     provide a config file for the registry server instance, default: ~/.config/veritas/config.yaml
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

package main

import (
	"context"

	"charm.land/log/v2"
	"github.com/j4ck4L0up3/veritas/internal/cli"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		log.Fatalf("Error: %v\n", err)
	}
}

func run(ctx context.Context) error {
	rootCmd := cli.NewCommand("0.0.1")
	cli.SetupCommands(rootCmd, ctx)

	if err := rootCmd.Execute(); err != nil {
		return err
	}

	return nil
}

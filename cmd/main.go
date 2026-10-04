package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"charm.land/log/v2"
	"github.com/j4ck4L0up3/veritas/db"
	"github.com/j4ck4L0up3/veritas/internal/cli"
)

var dbConn *sql.DB

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	dsn := fmt.Sprintf(
		"file:%s?_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000",
		"data/dev.db", // TODO: set db to prod db loc after testing
	)

	conn, err := db.RunMigrations(dsn)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error opening db: %v", err))
	}

	dbConn = conn
}

func run(ctx context.Context) error {
	rootCmd := cli.NewCommand("0.0.1")
	cli.SetupCommands(rootCmd, ctx, dbConn)

	if err := rootCmd.Execute(); err != nil {
		return err
	}

	return nil
}

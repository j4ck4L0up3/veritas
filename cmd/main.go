package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/j4ck4L0up3/veritas/internal/cli"
)

func main() {
	mux := http.NewServeMux()
	srvr := http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: mux,
	}

	mux.Handle(
		"POST /cool",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`you're here!`))
			fmt.Println(r.FormValue("my_query"))
		}),
	)

	srvr.ListenAndServe()

	/* ctx := context.Background()
	if err := run(ctx); err != nil {
		log.Fatalf("Error: %v\n", err)
	} */
}

func run(ctx context.Context) error {
	rootCmd := cli.NewCommand("0.0.1")
	cli.SetupCommands(rootCmd, ctx)

	if err := rootCmd.Execute(); err != nil {
		return err
	}

	return nil
}

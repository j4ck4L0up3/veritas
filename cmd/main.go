package main

import (
	"fmt"

	"github.com/j4ck4L0up3/veritas/internal/config"
)

func main() {
	cfg := config.Load()

	fmt.Printf("%+v", cfg)
}

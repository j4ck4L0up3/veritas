package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/j4ck4L0up3/veritas/internal/config"
	"github.com/j4ck4L0up3/veritas/internal/logger"
)

var cfg *config.Config
var lgr logger.Logger
var logFile *os.File

func main() {
	defer logFile.Close()
	lgr.Print("Starting Veritas...")
	lgr.Infof("Config loaded: %+v", cfg)
}

func init() {
	cfg = config.Load()

	if err := os.MkdirAll(cfg.Locations.LogPath, os.FileMode(0o700)); err != nil {
		panic(fmt.Sprintf("Error creating log directory: %v", err))
	}

	path := filepath.Join(cfg.Locations.LogPath, "veritas.log")
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if errors.Is(err, os.ErrExist) {
	} else if err != nil {
		panic(fmt.Sprintf("Error opening log file: %v", err))
	}

	logFile = f
	lgr = logger.New(logFile, cfg.LogLevel, cfg.LogFormat)
}

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"charm.land/log/v2"
)

const defaultYamlConfig = `
---
host: "127.0.0.1"
port: 9001
upload_ttl: 86_400 # in seconds; default = 24 hours
log_level: "info"
log_format: "text"
locations:
  db: "$HOME/.local/share/veritas"
  bin: "$HOME/.local/bin/veritas"
  env: "$HOME/.local/share/veritas"
	logs: "$HOME/.local/share/veritas" # optionally use "stdout" or "stderr"
  blobs: "$HOME/.local/share/veritas/blobs"
  uploads: "$HOME/.local/share/veritas/uploads"
  service: "$HOME/.config/systemd/user"
`

// instantiate yaml config to config path if not exists
func SetYamlConfig(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("error creating config dir %s: %w", path, err)
	}

	file := filepath.Join(path, "config.yaml")

	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		log.Info("config file already exists, skipping creation")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error creating config file: %w", err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Fatalf("unable to close config file: %v", err)
		}
	}()

	if _, err := f.WriteString(defaultYamlConfig); err != nil {
		return fmt.Errorf("error writing to config file: %w", err)
	}
	return nil
}

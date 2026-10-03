package config

import (
	"charm.land/log/v2"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const defaultConfigPath = ".config/veritas"

// get config path from env if exists, otherwise use default
func getConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("could not determine user home directory path")
	}

	if path := os.Getenv("VERITAS_CONFIG_PATH"); path != "" {
		return path // user should provide absolute path
	}

	return filepath.Join(home, defaultConfigPath)
}

const defaultYamlConfig = `
---
ip: "127.0.0.1"
port: 9001
upload_ttl: 86400 # in seconds; default = 24 hours
locations:
  db: ".local/share/veritas"
  bin: ".local/bin/veritas"
  logs: ".local/share/veritas"
  blobs: ".local/share/veritas/blobs"
  uploads: ".local/share/veritas/uploads"
  service: ".config/systemd/user"
`

// instantiate yaml config to config path if not exists
func setYamlConfig() error {
	path := getConfigPath()
	if path == "" {
		return errors.New("no config path found")
	}

	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("error creating config dir %s: %w", path, err)
	}

	file := filepath.Join(path, "config.yaml")

	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
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

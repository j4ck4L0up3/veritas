package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const defaultConfigPath = ".config/veritas"

// get config path from env if exists, otherwise use default
func getConfigPath() string {
	if path := os.Getenv("VERITAS_CONFIG_PATH"); path != "" {
		return path
	}

	if home := os.Getenv("HOME"); home != "" {
		return fmt.Sprintf("%s/%s", home, defaultConfigPath)
	}

	return ""
}

const defaultYamlConfig = `
---
  ip: "127.0.0.1"
  port: ":9001"
  locations:
    db: "~/.local/share/veritas"
    bin: "/usr/local/bin/veritas"
    blobs: "~/.local/share/veritas/blobs"
    uploads: "~/.local/share/veritas/uploads"
  upload_ttl: 86400 # 24 hours
`

// instantiate yaml config to config path if not exists
func setYamlConfig() error {
	path := getConfigPath()
	if path == "" {
		return errors.New("no config path found")
	}

	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("create config dir %s: %w", path, err)
	}

	file := filepath.Join(path, "config.yaml")

	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("create config file: %w", err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	if _, err := f.WriteString(defaultYamlConfig); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}

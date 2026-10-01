package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var config Config

type Config struct {
	IP        string    `yaml:"ip"`
	Port      string    `yaml:"port"`
	UploadTTL string    `yaml:"upload_ttl"`
	Locations Locations `yaml:"locations"`
}

type Locations struct {
	DbPath     string `yaml:"db"`
	BinPath    string `yaml:"bin"`
	BlobPath   string `yaml:"blobs"`
	UploadPath string `yaml:"uploads"`
}

func Load() *Config {
	must(load)
	setOverrides()
	return &config
}

func load() error {
	must(setYamlConfig)
	path := filepath.Join(getConfigPath(), "config.yaml")

	fmt.Println(path)
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return err
	}

	return nil
}

func must(f func() error) {
	if err := f(); err != nil {
		panic(err)
	}
}

const (
	VERITAS_DB_PATH     = "VERITAS_DB_PATH"
	VERITAS_BIN_PATH    = "VERITAS_BIN_PATH"
	VERITAS_BLOB_PATH   = "VERITAS_BLOB_PATH"
	VERITAS_UPLOAD_PATH = "VERITAS_UPLOAD_PATH"
	VERITAS_UPLOAD_TTL  = "VERITAS_UPLOAD_TTL"
	VERITAS_IP          = "VERITAS_IP"
	VERITAS_PORT        = "VERITAS_PORT"
)

func setOverrides() {
	if env := os.Getenv(VERITAS_DB_PATH); env != "" {
		config.Locations.DbPath = env
	}
	if env := os.Getenv(VERITAS_BIN_PATH); env != "" {
		config.Locations.BinPath = env
	}
	if env := os.Getenv(VERITAS_BLOB_PATH); env != "" {
		config.Locations.BlobPath = env
	}
	if env := os.Getenv(VERITAS_UPLOAD_PATH); env != "" {
		config.Locations.UploadPath = env
	}
	if env := os.Getenv(VERITAS_UPLOAD_TTL); env != "" {
		config.UploadTTL = env
	}
	if env := os.Getenv(VERITAS_IP); env != "" {
		config.IP = env
	}
	if env := os.Getenv(VERITAS_PORT); env != "" {
		config.Port = env
	}
}

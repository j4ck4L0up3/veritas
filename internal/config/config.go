package config

import (
	"charm.land/log/v2"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"strconv"
)

var config Config

type Config struct {
	IP        string    `yaml:"ip"`
	Port      uint      `yaml:"port"`
	UploadTTL string    `yaml:"upload_ttl"`
	LogLevel  string    `yaml:"log_level"`
	LogFormat string    `yaml:"log_format"`
	Locations Locations `yaml:"locations"`
}

type Locations struct {
	DbPath      string `yaml:"db"`
	BinPath     string `yaml:"bin"`
	LogPath     string `yaml:"logs"`
	BlobPath    string `yaml:"blobs"`
	UploadPath  string `yaml:"uploads"`
	ServicePath string `yaml:"service"`
}

func Load() *Config {
	must(load)
	setOverrides()
	return &config
}

func load() error {
	must(setYamlConfig)
	path := filepath.Join(getConfigPath(), "config.yaml")

	yamlBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// validate yaml
	schema, err := compileSchema()
	if err != nil {
		return err
	}

	if err := validate(schema, yamlBytes); err != nil {
		return err
	}

	err = yaml.Unmarshal(yamlBytes, &config)
	if err != nil {
		return err
	}

	// check for env vars
	config.Locations.BinPath = parseEnvPath(config.Locations.BinPath)
	config.Locations.DbPath = parseEnvPath(config.Locations.DbPath)
	config.Locations.LogPath = parseEnvPath(config.Locations.LogPath)
	config.Locations.BlobPath = parseEnvPath(config.Locations.BlobPath)
	config.Locations.UploadPath = parseEnvPath(config.Locations.UploadPath)
	config.Locations.ServicePath = parseEnvPath(config.Locations.ServicePath)

	return nil
}

func must(f func() error) {
	if err := f(); err != nil {
		log.Fatal(err)
	}
}

const (
	VERITAS_DB_PATH      = "VERITAS_DB_PATH"
	VERITAS_BIN_PATH     = "VERITAS_BIN_PATH"
	VERITAS_LOG_PATH     = "VERITAS_LOG_PATH"
	VERITAS_BLOB_PATH    = "VERITAS_BLOB_PATH"
	VERITAS_UPLOAD_PATH  = "VERITAS_UPLOAD_PATH"
	VERITAS_SERVICE_PATH = "VERITAS_SERVICE_PATH"
	VERITAS_UPLOAD_TTL   = "VERITAS_UPLOAD_TTL"
	VERITAS_IP           = "VERITAS_IP"
	VERITAS_PORT         = "VERITAS_PORT"
	VERITAS_LOG_LEVEL    = "VERITAS_LOG_LEVEL"
	VERITAS_LOG_FORMAT   = "VERITAS_LOG_FORMAT"
)

func setOverrides() {
	if env := os.Getenv(VERITAS_DB_PATH); env != "" {
		config.Locations.DbPath = env
	}
	if env := os.Getenv(VERITAS_BIN_PATH); env != "" {
		config.Locations.BinPath = env
	}
	if env := os.Getenv(VERITAS_LOG_PATH); env != "" {
		config.Locations.LogPath = env
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
		port, err := strconv.Atoi(env)
		if err != nil {
			log.Fatalf("non-integer port assigned to VERITAS_PORT: %v", err)
		}

		config.Port = verifiedPort(port)
	}
}

func verifiedPort(port int) uint {
	if port < 0 || port > 65535 {
		log.Fatal("invalid port")
	}

	if port < 1024 {
		log.Warn("WARNING: port less than 1024, may conflict with existing services")
	}

	return uint(port)
}

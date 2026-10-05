package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"charm.land/log/v2"

	"strconv"

	"gopkg.in/yaml.v3"
)

var config Config

type Config struct {
	Host      string    `yaml:"host"`
	Port      uint16    `yaml:"port"`
	UploadTTL uint64    `yaml:"upload_ttl"`
	LogLevel  string    `yaml:"log_level"`
	LogFormat string    `yaml:"log_format"`
	Locations Locations `yaml:"locations"`
}

type Locations struct {
	DbPath      string `yaml:"db"`
	BinPath     string `yaml:"bin"`
	EnvPath     string `yaml:"env"`
	LogPath     string `yaml:"logs"`
	BlobPath    string `yaml:"blobs"`
	UploadPath  string `yaml:"uploads"`
	ServicePath string `yaml:"service"`
}

func Load(cmdPath string) (*Config, error) {
	var path string
	if cmdPath == "" {
		path = filepath.Join(getConfigPath(), "config.yaml")
		if err := setYamlConfig(getConfigPath()); err != nil {
			return nil, err
		}
	} else {
		path = cmdPath
	}

	if err := load(path); err != nil {
		return nil, err
	}

	if err := setOverrides(); err != nil {
		return nil, err
	}

	return &config, nil
}

func load(path string) error {
	yamlBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// validate config file
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

	// check for env path vars
	config.Locations.DbPath = parseEnvPath(config.Locations.DbPath)
	config.Locations.BinPath = parseEnvPath(config.Locations.BinPath)
	config.Locations.EnvPath = parseEnvPath(config.Locations.EnvPath)
	config.Locations.LogPath = parseEnvPath(config.Locations.LogPath)
	config.Locations.BlobPath = parseEnvPath(config.Locations.BlobPath)
	config.Locations.UploadPath = parseEnvPath(config.Locations.UploadPath)
	config.Locations.ServicePath = parseEnvPath(config.Locations.ServicePath)

	// verify port
	validPort, err := verifiedPort(int(config.Port))
	if errors.Is(err, ErrConflictingPort) {
		log.Warn(err.Error())
	} else if err != nil {
		return err
	}

	config.Port = validPort

	return nil
}

const (
	VERITAS_DB_PATH      = "VERITAS_DB_PATH"
	VERITAS_BIN_PATH     = "VERITAS_BIN_PATH"
	VERITAS_ENV_PATH     = "VERITAS_ENV_PATH"
	VERITAS_LOG_PATH     = "VERITAS_LOG_PATH"
	VERITAS_BLOB_PATH    = "VERITAS_BLOB_PATH"
	VERITAS_UPLOAD_PATH  = "VERITAS_UPLOAD_PATH"
	VERITAS_SERVICE_PATH = "VERITAS_SERVICE_PATH"
	VERITAS_UPLOAD_TTL   = "VERITAS_UPLOAD_TTL"
	VERITAS_HOST         = "VERITAS_HOST"
	VERITAS_PORT         = "VERITAS_PORT"
	VERITAS_LOG_LEVEL    = "VERITAS_LOG_LEVEL"
	VERITAS_LOG_FORMAT   = "VERITAS_LOG_FORMAT"
)

func setOverrides() error {
	if env := os.Getenv(VERITAS_DB_PATH); env != "" {
		config.Locations.DbPath = parseEnvPath(env)
	}
	if env := os.Getenv(VERITAS_BIN_PATH); env != "" {
		config.Locations.BinPath = parseEnvPath(env)
	}
	if env := os.Getenv(VERITAS_ENV_PATH); env != "" {
		config.Locations.EnvPath = parseEnvPath(env)
	}
	if env := os.Getenv(VERITAS_LOG_PATH); env != "" {
		config.Locations.LogPath = parseEnvPath(env)
	}
	if env := os.Getenv(VERITAS_BLOB_PATH); env != "" {
		config.Locations.BlobPath = parseEnvPath(env)
	}
	if env := os.Getenv(VERITAS_UPLOAD_PATH); env != "" {
		config.Locations.UploadPath = parseEnvPath(env)
	}
	if env := os.Getenv(VERITAS_SERVICE_PATH); env != "" {
		config.Locations.ServicePath = parseEnvPath(env)
	}
	if env := os.Getenv(VERITAS_HOST); env != "" {
		config.Host = parseEnvPath(env)
	}
	if env := os.Getenv(VERITAS_PORT); env != "" {
		port, err := strconv.Atoi(env)
		if err != nil {
			log.Fatalf("non-integer port assigned to VERITAS_PORT: %v", err)
		}

		validPort, err := verifiedPort(port)
		if errors.Is(err, ErrConflictingPort) {
			log.Warn(err.Error())
		} else if err != nil {
			return err
		}

		config.Port = validPort
	}
	if env := os.Getenv(VERITAS_UPLOAD_TTL); env != "" {
		ttl, err := strconv.Atoi(env)
		if err != nil {
			return fmt.Errorf("non-integer ttl assigned to VERITAS_UPLOAD_TTL: %v", err)
		}

		config.UploadTTL = uint64(ttl)
	}

	return nil
}

var ErrConflictingPort = errors.New("port less than 1024, may conflict with existing services")
var ErrInvalidPort = errors.New("invalid port")

func verifiedPort(port int) (uint16, error) {
	if port < 0 || port > 65535 {
		return 0, ErrInvalidPort
	}

	if port < 1024 {
		return uint16(port), ErrConflictingPort
	}

	return uint16(port), nil
}

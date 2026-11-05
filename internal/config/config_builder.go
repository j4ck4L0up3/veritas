package config

import (
	"charm.land/log/v2"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
	"strings"
)

const (
	defaultConfigFile    = "$HOME/.config/veritas/config.yaml"
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

type ConfigBuilder struct {
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

func newConfigBuilder() *ConfigBuilder {
	locations := Locations{
		DbPath:      "$HOME/.local/share/veritas",
		BinPath:     "$HOME/.local/bin/veritas",
		EnvPath:     "$HOME/.local/share/veritas",
		LogPath:     "$HOME/.local/share/veritas",
		BlobPath:    "$HOME/.local/share/veritas/blobs",
		UploadPath:  "$HOME/.local/share/veritas/uploads",
		ServicePath: "$HOME/.config/systemd/user",
	}

	return &ConfigBuilder{
		Host:      "127.0.0.1",
		Port:      9001,
		UploadTTL: 86_400,
		LogLevel:  "info",
		LogFormat: "text",
		Locations: locations,
	}
}

func (c *ConfigBuilder) build() *Config {
	return &Config{ConfigBuilder: *c}
}

func (c *ConfigBuilder) loadYamlConfigFile(path string) error {
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

	err = yaml.Unmarshal(yamlBytes, c)
	if err != nil {
		return err
	}

	// check for env path vars
	c.Locations.DbPath = parseEnvPath(c.Locations.DbPath)
	c.Locations.BinPath = parseEnvPath(c.Locations.BinPath)
	c.Locations.EnvPath = parseEnvPath(c.Locations.EnvPath)
	c.Locations.LogPath = parseEnvPath(c.Locations.LogPath)
	c.Locations.BlobPath = parseEnvPath(c.Locations.BlobPath)
	c.Locations.UploadPath = parseEnvPath(c.Locations.UploadPath)
	c.Locations.ServicePath = parseEnvPath(c.Locations.ServicePath)

	// verify port
	validPort, err := verifiedPort(int(c.Port))
	if errors.Is(err, ErrConflictingPort) {
		log.Warn(err.Error())
	} else if err != nil {
		return err
	}

	c.Port = validPort

	return nil
}

func (c *ConfigBuilder) setEnvOverrides() error {
	if env := os.Getenv(VERITAS_DB_PATH); env != "" {
		c.Locations.DbPath = parseEnvPath(env)
	}
	if env := os.Getenv(VERITAS_BIN_PATH); env != "" {
		c.Locations.BinPath = parseEnvPath(env)
	}
	if env := os.Getenv(VERITAS_ENV_PATH); env != "" {
		c.Locations.EnvPath = parseEnvPath(env)
	}
	if env := os.Getenv(VERITAS_LOG_PATH); env != "" {
		c.Locations.LogPath = parseEnvPath(env)
	}
	if env := os.Getenv(VERITAS_BLOB_PATH); env != "" {
		c.Locations.BlobPath = parseEnvPath(env)
	}
	if env := os.Getenv(VERITAS_UPLOAD_PATH); env != "" {
		c.Locations.UploadPath = parseEnvPath(env)
	}
	if env := os.Getenv(VERITAS_SERVICE_PATH); env != "" {
		c.Locations.ServicePath = parseEnvPath(env)
	}
	if env := os.Getenv(VERITAS_HOST); env != "" {
		c.Host = parseEnvPath(env)
	}
	if env := os.Getenv(VERITAS_PORT); env != "" {
		port, err := strconv.Atoi(env)
		if err != nil {
			return fmt.Errorf("non-integer port assigned to VERITAS_PORT: %v", err)
		}

		validPort, err := verifiedPort(port)
		if errors.Is(err, ErrConflictingPort) {
			log.Warn(err.Error())
		} else if err != nil {
			return err
		}

		c.Port = validPort
	}
	if env := os.Getenv(VERITAS_UPLOAD_TTL); env != "" {
		ttl, err := strconv.Atoi(env)
		if err != nil {
			return fmt.Errorf("non-integer ttl assigned to VERITAS_UPLOAD_TTL: %v", err)
		}

		c.UploadTTL = uint64(ttl)
	}

	return nil
}

// TODO: add checks for these errors
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

func getConfigPath() string {
	if path := os.Getenv("VERITAS_CONFIG_PATH"); path != "" {
		return parseEnvPath(path)
	}

	if configPath := os.Getenv("XDG_CONFIG_HOME"); configPath != "" {
		return parseEnvPath(fmt.Sprintf("%s/veritas/config.yaml", configPath))
	}

	return parseEnvPath(defaultConfigFile)
}

func parseEnvPath(path string) string {
	if !strings.Contains(path, "$") {
		return path
	}

	prependSlash := path[0] == '/'
	parts := strings.Split(path, "/")

	for i := range parts {
		if strings.HasPrefix(parts[i], "$") {
			var ok bool
			parts[i], ok = os.LookupEnv(strings.TrimPrefix(parts[i], "$"))
			if !ok {
				// load() attempts to read it and it will fail there with file not exists
				return path
			}
		}
	}

	if prependSlash {
		return "/" + filepath.Join(parts...)
	}

	return filepath.Join(parts...)
}

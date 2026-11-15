package config

import (
	"charm.land/log/v2"
	"errors"
	"github.com/j4ck4L0up3/veritas/internal/logger"
	"os"
	"path/filepath"
)

type Config struct {
	ConfigBuilder
}

// load config with predence of flags > env > yaml file > default
func Load(flagConfig string) *Config {
	var config *Config
	builder := newConfigBuilder()
	log.Debug("config flag", flagConfig)

	if flagConfig == "" {
		configFile := getConfigPath()
		log.Debug("configFile", configFile)
		err := builder.loadYamlConfigFile(configFile)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				log.Warnf("config file does not exist at %s", configFile)
			} else {
				log.Fatalf("failed to load config file at %s: %v", configFile, err)
			}
		}
	}

	if err := builder.setEnvOverrides(); err != nil {
		log.Fatalf("failed to set environment overrides: %v", err)
	}

	if flagConfig != "" {
		if err := builder.loadYamlConfigFile(flagConfig); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				log.Warnf("config file does not exist at %s", flagConfig)
			} else {
				log.Fatalf("failed to load config file at %s: %v", flagConfig, err)
			}
		}
	}

	config = builder.build()

	return config
}

func (cfg *Config) InitLogger(path string) logger.Logger {
	var lgr logger.Logger
	var logPath string

	switch path {
	case "stdout":
		lgr = logger.New(os.Stdout, cfg.LogLevel, cfg.LogFormat)
		return lgr
	case "stderr":
		lgr = logger.New(os.Stderr, cfg.LogLevel, cfg.LogFormat)
		return lgr
	case "":
		switch cfg.Locations.LogPath {
		case "stdout":
			lgr = logger.New(os.Stdout, cfg.LogLevel, cfg.LogFormat)
			return lgr
		case "stderr":
			lgr = logger.New(os.Stderr, cfg.LogLevel, cfg.LogFormat)
			return lgr
		default:
			logPath = filepath.Join(cfg.Locations.LogPath, "veritas.log")
		}
	default:
		logPath = path
	}

	logFile, err := os.OpenFile(
		logPath,
		os.O_WRONLY|os.O_CREATE,
		0o200,
	)

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("something went wrong while initializing logger from log file: %s", err)
	}

	lgr = logger.New(logFile, cfg.LogLevel, cfg.LogFormat)

	return lgr
}

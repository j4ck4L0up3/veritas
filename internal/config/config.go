package config

import (
	"charm.land/log/v2"
	"errors"
	"os"
)

type Config struct {
	ConfigBuilder
}

// load config with predence of flags > env > yaml file > default
func Load(flagConfig string) *Config {
	var config *Config
	builder := newConfigBuilder()

	if flagConfig == "" {
		configFile := getConfigPath()
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

	if err := builder.loadYamlConfigFile(flagConfig); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Warnf("config file does not exist at %s", flagConfig)
		} else {
			log.Fatalf("failed to load config file at %s: %v", flagConfig, err)
		}
	}

	config = builder.build()

	return config
}

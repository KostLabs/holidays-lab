package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type MongoDB struct {
	URI        string `yaml:"uri"`
	Database   string `yaml:"database"`
	Collection string `yaml:"collection"`
}

type ExternalAPI struct {
	URL string `yaml:"url"`
}

type Config struct {
	Port        int         `yaml:"port"`
	BasePath    string      `yaml:"base_path"`
	MongoDB     MongoDB     `yaml:"mongodb"`
	ExternalAPI ExternalAPI `yaml:"external_api"`
}

func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	unmarshalErr := yaml.Unmarshal(data, &cfg)
	if unmarshalErr != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", unmarshalErr)
	}

	return &cfg, nil
}

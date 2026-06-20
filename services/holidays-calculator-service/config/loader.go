package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type HolidaysAPIService struct {
	URL string `yaml:"url"`
}

type Config struct {
	Port               int                `yaml:"port"`
	BasePath           string             `yaml:"base_path"`
	HolidaysAPIService HolidaysAPIService `yaml:"holidays_api_service"`
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

package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type External struct {
	Name string `yaml:"name"`
	Port int    `yaml:"port"`
	URL  string `yaml:"url"`
}

type Config struct {
	Port      int        `yaml:"port"`
	BasePath  string     `yaml:"base_path"`
	Externals []External `yaml:"externals"`
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

func (c *Config) GetExternalByName(name string) (*External, bool) {
	for _, ext := range c.Externals {
		if ext.Name == name {
			return &ext, true
		}
	}
	return nil, false
}

package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Zinc is the configuration for the ZincSearch server
type Zinc struct {
	URL string `yaml:"URL"`
}

// ServerConfig is the config for the server
type ServerConfig struct {
	Port string `yaml:"port"`
}

// Config contains the configuration for the server
type Config struct {
	Server ServerConfig `yaml:"Server"`
}

// LoadConfig loads the config file
func LoadConfig(filename string) (*Config, error) {
	config := &Config{}

	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(content, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

package config

import (
	"flag"
	"os"

	"gopkg.in/yaml.v2"
)

// Zinc is the configuration for the ZincSearch server
type Zinc struct {
	ZincPort string `yaml:"zincPort"`
	Target   string `yaml:"target"`
}

// ServerConfig is the config for the server
type ServerConfig struct {
	Port string `yaml:"port"`
}

// Config contains the configuration for the server
type Config struct {
	Server ServerConfig `yaml:"Server"`
	Zinc   Zinc         `yaml:"Zinc"`
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

	config.Server.Port = *flag.String("port", config.Server.Port, "server port")
	config.Zinc.ZincPort = *flag.String("zincPort", config.Zinc.ZincPort, "zinc port")
	config.Zinc.Target = *flag.String("target", config.Zinc.Target, "target")
	flag.Parse()

	return config, nil
}

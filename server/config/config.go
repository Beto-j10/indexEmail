package config

import (
	"errors"
	"flag"
	"os"

	"gopkg.in/yaml.v2"
)

// Zinc is the configuration for the ZincSearch server
type Zinc struct {
	ZincHost  string `yaml:"zincHost"`
	Target    string `yaml:"target"`
	DocCreate string `yaml:"doc_create"`
	User      string `yaml:"user"`
	Password  string `yaml:"password"`
}

// ServerConfig is the config for the server
type ServerConfig struct {
	Port string `yaml:"port"`
	Dir  string `yaml:"dir"`
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

	serverPort := flag.String("port", config.Server.Port, "server port")
	serverDir := flag.String("dir", "default", "server directory")
	zincHost := flag.String("zincHost", config.Zinc.ZincHost, "zinc host")
	zincTarget := flag.String("target", config.Zinc.Target, "target")

	flag.Parse()

	config.Server.Port = *serverPort
	config.Server.Dir = *serverDir
	config.Zinc.ZincHost = *zincHost
	config.Zinc.Target = *zincTarget

	if config.Server.Dir == "default" {
		err := errors.New("please specify a directory to index: -dir=<path>")
		return nil, err
	}

	return config, nil
}

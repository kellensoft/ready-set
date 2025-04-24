package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Apps []App `yaml:"apps"`
}

type App struct {
	Name     string            `yaml:"name"`     // directory name under /apps/
	Repo     string            `yaml:"repo"`     // Git repository URL
	Env      map[string]string `yaml:"env"`      // Environment variables for this app
	Commands map[string]string `yaml:"commands"` // build, test, start commands
}

// LoadConfig reads and parses the config.yaml file
func LoadConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config AppConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

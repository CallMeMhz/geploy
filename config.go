package geploy

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type DeployConfig struct {
	Service     string   `yaml:"service"`
	Port        string   `yaml:"port"`
	Image       string   `yaml:"image"`
	Servers     []string `yaml:"servers"`
	HealthCheck string   `yaml:"health_check,omitempty"`

	SshConfig struct {
		Username string `yaml:"usr"`
	} `yaml:"ssh,omitempty"`

	Registry struct {
		Username string `yaml:"usr"`
		Password string `yaml:"pwd"`
	} `yaml:"registry"`
}

func LoadDeployConfig() (*DeployConfig, error) {
	workdir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filepath.Join(workdir, "deploy.yml"))
	if err != nil {
		return nil, err
	}

	cfg := new(DeployConfig)
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.HealthCheck == "" {
		cfg.HealthCheck = "/ping"
	}
	if cfg.SshConfig.Username == "" {
		cfg.SshConfig.Username = "root"
	}
	return cfg, nil
}

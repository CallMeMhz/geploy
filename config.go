package geploy

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
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

	Secret []string `yaml:"secret,omitempty"`
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

	if len(cfg.Secret) > 0 {
		secret := cfg.Secret
		if err := godotenv.Load(); err != nil {
			return nil, err
		}
		for _, field := range cfg.Secret {
			data = bytes.ReplaceAll(data, []byte(field), []byte(os.Getenv(field)))
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
		cfg.Secret = secret
	}

	if cfg.HealthCheck == "" {
		cfg.HealthCheck = "/ping"
	}
	if cfg.SshConfig.Username == "" {
		cfg.SshConfig.Username = "root"
	}
	return cfg, nil
}

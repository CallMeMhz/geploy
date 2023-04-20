package geploy

import (
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
		Username any `yaml:"usr"`
		Password any `yaml:"pwd"`
	} `yaml:"registry"`
}

func LoadDeployConfig() (*DeployConfig, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

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

func lookup(key any) string {
	switch key := key.(type) {
	case string:
		return key
	case []any:
		return os.Getenv(key[0].(string))
	default:
		panic("key not found")
	}
}

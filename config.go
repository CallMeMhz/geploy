package geploy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
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
		// todo customize registry center
		Username any `yaml:"usr"`
		Password any `yaml:"pwd"`
	} `yaml:"registry"`

	Env []string `yaml:"env"`
	env map[string]string
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
	cfg.env = make(map[string]string)

	for _, env := range cfg.Env {
		var key, value string
		if strings.Contains(env, "=") {
			// set env literally
			if parts := strings.SplitN(env, "=", 2); len(parts) == 1 {
				// in case of no value
				key = parts[0]
			} else {
				key, value = parts[0], parts[1]
			}
		} else {
			// inherit from current environment
			key, value = env, os.Getenv(env)
		}
		if value != "" {
			cfg.env[key] = value
		} else {
			fmt.Println("Environment variable", color.HiYellowString(key), "is missing")
		}
	}

	// set default values
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

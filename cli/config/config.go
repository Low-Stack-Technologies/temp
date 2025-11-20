package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the CLI configuration
type Config struct {
	Server string `yaml:"server"`
}

// Load loads the configuration from file and environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: "https://temp.low-stack.tech",
	}

	// Load from config file
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(homeDir, ".config", "temp.yaml")
		if _, err := os.Stat(configPath); err == nil {
			data, err := os.ReadFile(configPath)
			if err == nil {
				var fileCfg Config
				if err := yaml.Unmarshal(data, &fileCfg); err == nil {
					if fileCfg.Server != "" {
						cfg.Server = fileCfg.Server
					}
				}
			}
		}
	}

	// Override with environment variable
	if envServer := os.Getenv("TEMP_SERVER"); envServer != "" {
		cfg.Server = envServer
	}

	return cfg, nil
}

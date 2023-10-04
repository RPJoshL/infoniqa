package models

import (
	"os"

	"git.rpjosh.de/RPJosh/go-logger"
	"gitea.hama.de/LFS/infoniqa-scripts/pkg/utils"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Url      string `yaml:"url"`
}

// GetConfig reads the configuration from the file and returns this config struct.
// The program will be left if no valid configuration file was found
func GetConfig() *Config {
	rtc := Config{}

	// Get the configuration path
	configPath := utils.GetEnvString("INFONIQA_CONFIG", "./config")

	// Parse the file
	dat, err := os.ReadFile(configPath)
	if err != nil {
		logger.Fatal("Failed to read configuration file %q: %s", configPath, err)
	}

	// Unmarshal
	if err := yaml.Unmarshal(dat, &rtc); err != nil {
		logger.Fatal("Failed to unmarshal configuration file: %s", err)
	}

	return &rtc
}

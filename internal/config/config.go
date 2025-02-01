package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Tagger     []TagConfig  `yaml:"tagger"`
	Formatting []Formatter  `yaml:"formatting"`
	LLM        LLMConfig    `yaml:"llm"`
}

type LLMConfig struct {
	DefaultProvider string                 `yaml:"default_provider"`
	Providers       map[string]ProviderConfig `yaml:"providers"`
}

type ProviderConfig struct {
	Type   string                 `yaml:"type"`
	Model  string                 `yaml:"model"`
	Params map[string]interface{} `yaml:"params"`
}

func Load() (Config, error) {
	config_filename, err := getConfigFile("acme-utils", "config.yaml")
	if err != nil {
		return Config{}, fmt.Errorf("could not get configuration file path", err)
	}

	f, err := os.Open(config_filename)
	if err != nil {
		return Config{},
			fmt.Errorf("could not open configuration file: %v", err)
	}

	c := Config{}
	d := yaml.NewDecoder(f)
	err = d.Decode(&c)
	if err != nil {
		return Config{},
			fmt.Errorf("could not decode the configuration file: %v", err)
	}

	return c, nil
}

// getConfigFile constructs the full path for a given application's config file.
func getConfigFile(appName, fileName string) (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}

	// Construct the full path to the configuration file.
	filePath := filepath.Join(configDir, appName, fileName)
	return filePath, nil
}

// getConfigDir retrieves the appropriate config directory for the current user.
func getConfigDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	// Use the XDG_CONFIG_HOME environment variable if set,
	// otherwise fallback to the default location.
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(usr.HomeDir, ".config")
	}

	return configHome, nil
}

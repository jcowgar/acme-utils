package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type TagConfiguration struct {
	Extension string `yaml:"extension"`
	CheckTag  string `yaml:"checkTag"`
	Tag       string `yaml:"tag"`
}

type OllamaConfiguration struct {
	DefaultModel string `yaml:"default_model"`
	BaseURL      string `yaml:"base_url"`
}

type Configuration struct {
	Tagger []TagConfiguration  `yaml:"tagger"`
	Ollama OllamaConfiguration `yaml:"ollama"`
}

func Load() (Configuration, error) {
	config_filename, err := getConfigFile("acme-utils", "config.yaml")
	if err != nil {
		return Configuration{}, fmt.Errorf("could not get configuration file path", err)
	}

	f, err := os.Open(config_filename)
	if err != nil {
		return Configuration{},
			fmt.Errorf("could not open configuration file: %v", err)
	}

	c := Configuration{}
	d := yaml.NewDecoder(f)
	err = d.Decode(&c)
	if err != nil {
		return Configuration{},
			fmt.Errorf("could not decode the configuration file: %v", err)
	}

	return c, nil
}

func (c Configuration) TaggerFor(extension string) *TagConfiguration {
	extension = strings.TrimPrefix(extension, ".")
	for _, v := range c.Tagger {
		if v.Extension == extension {
			return &v
		}
	}

	return nil
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

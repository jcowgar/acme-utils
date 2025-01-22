package tagger

import (
	"fmt"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type ConditionConfiguration struct {
	ParentDirectory string `yaml:"parentDirectory"`
}

type TagConfiguration struct {
	Extension string `yaml:"extension"`
	CheckTag  string `yaml:"checkTag"`
	Tag       string `yaml:"tag"`
}

type Configuration []TagConfiguration

func LoadConfiguration() (Configuration, error) {
	f, err := os.Open("/home/jeremy/.local/acme/go/tagger.yaml")
	if err != nil {
		return Configuration{},
			fmt.Errorf("could not open tagger.yaml configuration file: %v", err)
	}

	c := Configuration{}
	d := yaml.NewDecoder(f)
	err = d.Decode(&c)
	if err != nil {
		return Configuration{},
			fmt.Errorf("could not decode tagger.yaml configuration file: %v", err)
	}

	return c, nil
}

func (c Configuration) Find(extension string) *TagConfiguration {
	extension = strings.TrimPrefix(extension, ".")
	for _, v := range c {
		if v.Extension == extension {
			return &v
		}
	}

	return nil
}

package config

import "strings"

type TagConfig struct {
	Extension string `yaml:"extension"`
	CheckTag  string `yaml:"checkTag"`
	Tag       string `yaml:"tag"`
}


func (c Config) TaggerFor(extension string) *TagConfig {
	extension = strings.TrimPrefix(extension, ".")
	for _, v := range c.Tagger {
		if v.Extension == extension {
			return &v
		}
	}

	return nil
}

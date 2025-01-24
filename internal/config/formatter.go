package config

import "strings"

type Formatter struct {
	Extension string `yaml:"extension"`
	Command   string `yaml:"command"`
	InPlace   bool   `yaml:"in_place"`
}

func (c Config) FormatterFor(extension string) *Formatter {
	extension = strings.TrimPrefix(extension, ".")
	for _, v := range c.Formatting {
		if v.Extension == extension {
			return &v
		}
	}

	return nil
}

package main

import (
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

func loadQueries(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	dec := yaml.NewDecoder(f)
	var config Config
	err = dec.Decode(&config)
	if err != nil {
		return nil, err
	}
	massageYaml(&config)
	return &config, nil
}

func massageYaml(c *Config) {
	for _, group := range c.Groups {
		for _, q := range group.Queries {
			// Remove newlines and extra space from "readable" yaml
			parts := strings.Split(q.Query, "\n")
			for i := range parts {
				parts[i] = strings.TrimSpace(parts[i])
			}
			q.Query = strings.TrimSpace(strings.Join(parts, " "))
		}
	}
}

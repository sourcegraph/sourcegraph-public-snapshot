package main

import (
	"os"

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
	return &config, err
}

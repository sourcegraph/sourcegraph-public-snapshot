package config

import (
	"encoding/json"
	"os"
)

type GitHub struct {
	Org      string `json:"Org"`
	URL      string `json:"URL"`
	User     string `json:"User"`
	Password string `json:"Password"`
	Token    string `json:"Token"`
}

type SourcegraphCfg struct {
	URL      string `json:"URL"`
	User     string `json:"User"`
	Password string `json:"Password"`
	Token    string `json:"Token"`
}

type Config struct {
	GitHub      GitHub         `json:"GitHub"`
	Sourcegraph SourcegraphCfg `json:"Sourcegraph"`
}

func FromFile(filename string) (*Config, error) {
	var c Config

	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	if err := json.NewDecoder(fd).Decode(&c); err != nil {
		return nil, err
	}

	return &c, nil
}

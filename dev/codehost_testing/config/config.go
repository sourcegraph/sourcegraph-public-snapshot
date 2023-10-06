package config

import (
	"encoding/json"
	"os"
)

type GitHub struct {
	URL       string `json:"url"`
	AdminUser string `json:"adminUser"`
	Password  string `json:"password"`
	Token     string `json:"token"`
}

type SourcegraphCfg struct {
	URL      string `json:"url"`
	User     string `json:"user"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

type Config struct {
	GitHub      GitHub         `json:"github"`
	Sourcegraph SourcegraphCfg `json:"sourcegraph"`
}

func FromFile(filename string) (*Config, error) {
	var c Config

	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	if err := json.NewDecoder(fd).Decode(&c); err != nil {
		return nil, err
	}

	return &c, nil
}

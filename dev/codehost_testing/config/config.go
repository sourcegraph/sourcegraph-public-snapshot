package config

import (
	"encoding/json"
	"os"
)

// GitHub represents the GitHub client configuration to connect  to a GitHub codehost
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

// Config represents the configuration which should be used when connecting to codehosts.
//
// Currently we only support connecting to GitHub.
type Config struct {
	GitHub      GitHub         `json:"github"`
	Sourcegraph SourcegraphCfg `json:"sourcegraph"`
}

// FromFile reads the configuration from the specified file. This method will return an error when we either fail
// to open a file or fail to decode the JSON into the Config struct.
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

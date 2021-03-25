package main

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func ParseConfigFile(name string) (*Config, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open file %q", name)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.Wrap(err, "reading configuration file")
	}

	var conf Config
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, err
	}

	for name, cmd := range conf.Commands {
		cmd.Name = name
		conf.Commands[name] = cmd
	}

	return &conf, nil
}

type Command struct {
	Name             string
	Cmd              string            `yaml:"cmd"`
	Install          string            `yaml:"install"`
	Env              map[string]string `yaml:"env"`
	Watch            []string          `yaml:"watch"`
	InstallDocDarwin string            `yaml:"install_doc.darwin"`
	InstallDocLinux  string            `yaml:"install_doc.linux"`
}

type Config struct {
	Env         map[string]string   `yaml:"env"`
	Commands    map[string]Command  `yaml:"commands"`
	Commandsets map[string][]string `yaml:"commandsets"`
	Tests       map[string]Command  `yaml:"tests"`
}

package main

//
// TODO: expose this directly in go-mockgen
//

import (
	"os"

	"gopkg.in/yaml.v3"
)

type yamlPayload struct {
	// Meta options
	IncludeConfigPaths []string `yaml:"include-config-paths,omitempty"`

	// Global options
	Exclude           []string `yaml:"exclude,omitempty"`
	Prefix            string   `yaml:"prefix,omitempty"`
	ConstructorPrefix string   `yaml:"constructor-prefix,omitempty"`
	Force             bool     `yaml:"force,omitempty"`
	DisableFormatting bool     `yaml:"disable-formatting,omitempty"`
	Goimports         string   `yaml:"goimports,omitempty"`
	ForTest           bool     `yaml:"for-test,omitempty"`
	FilePrefix        string   `yaml:"file-prefix,omitempty"`

	StdlibRoot string `yaml:"stdlib-root,omitempty"`

	Mocks []yamlMock `yaml:"mocks,omitempty"`
}

type yamlMock struct {
	Path              string       `yaml:"path,omitempty"`
	Paths             []string     `yaml:"paths,omitempty"`
	Sources           []yamlSource `yaml:"sources,omitempty"`
	SourceFiles       []string     `yaml:"source-files,omitempty"`
	Archives          []string     `yaml:"archives,omitempty"`
	Package           string       `yaml:"package,omitempty"`
	Interfaces        []string     `yaml:"interfaces,omitempty"`
	Exclude           []string     `yaml:"exclude,omitempty"`
	Dirname           string       `yaml:"dirname,omitempty"`
	Filename          string       `yaml:"filename,omitempty"`
	ImportPath        string       `yaml:"import-path,omitempty"`
	Prefix            string       `yaml:"prefix,omitempty"`
	ConstructorPrefix string       `yaml:"constructor-prefix,omitempty"`
	Force             bool         `yaml:"force,omitempty"`
	DisableFormatting bool         `yaml:"disable-formatting,omitempty"`
	Goimports         string       `yaml:"goimports,omitempty"`
	ForTest           bool         `yaml:"for-test,omitempty"`
	FilePrefix        string       `yaml:"file-prefix,omitempty"`
}

type yamlSource struct {
	Path        string   `yaml:"path,omitempty"`
	Paths       []string `yaml:"paths,omitempty"`
	Interfaces  []string `yaml:"interfaces,omitempty"`
	Exclude     []string `yaml:"exclude,omitempty"`
	Prefix      string   `yaml:"prefix,omitempty"`
	SourceFiles []string `yaml:"source-files,omitempty"`
}

func readManifest() (yamlPayload, error) {
	contents, err := os.ReadFile("mockgen.yaml")
	if err != nil {
		return yamlPayload{}, err
	}

	var payload yamlPayload
	if err := yaml.Unmarshal(contents, &payload); err != nil {
		return yamlPayload{}, err
	}

	for _, path := range payload.IncludeConfigPaths {
		payload, err = readIncludeConfig(payload, path)
		if err != nil {
			return yamlPayload{}, err
		}
	}

	return payload, nil
}

func readIncludeConfig(payload yamlPayload, path string) (yamlPayload, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return yamlPayload{}, err
	}

	var mocks []yamlMock
	if err := yaml.Unmarshal(contents, &mocks); err != nil {
		return yamlPayload{}, err
	}

	payload.Mocks = append(payload.Mocks, mocks...)
	return payload, nil
}

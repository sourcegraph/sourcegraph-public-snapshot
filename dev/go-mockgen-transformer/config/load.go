package config

//
// TODO: expose this directly in go-mockgen
//

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type YamlPayload struct {
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

	Mocks []YamlMock `yaml:"mocks,omitempty"`
}

type YamlMock struct {
	Path              string       `yaml:"path,omitempty"`
	Paths             []string     `yaml:"paths,omitempty"`
	Sources           []YamlSource `yaml:"sources,omitempty"`
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

type YamlSource struct {
	Path        string   `yaml:"path,omitempty"`
	Paths       []string `yaml:"paths,omitempty"`
	Interfaces  []string `yaml:"interfaces,omitempty"`
	Exclude     []string `yaml:"exclude,omitempty"`
	Prefix      string   `yaml:"prefix,omitempty"`
	SourceFiles []string `yaml:"source-files,omitempty"`
}

func ReadManifest(configPath string) (YamlPayload, error) {
	contents, err := os.ReadFile(configPath)
	if err != nil {
		return YamlPayload{}, err
	}

	var payload YamlPayload
	if err := yaml.Unmarshal(contents, &payload); err != nil {
		return YamlPayload{}, err
	}

	for _, path := range payload.IncludeConfigPaths {
		payload, err = readIncludeConfig(payload, filepath.Join(filepath.Dir(configPath), path))
		if err != nil {
			return YamlPayload{}, err
		}
	}

	return payload, nil
}

func readIncludeConfig(payload YamlPayload, path string) (YamlPayload, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return YamlPayload{}, err
	}

	var mocks []YamlMock
	if err := yaml.Unmarshal(contents, &mocks); err != nil {
		return YamlPayload{}, err
	}

	payload.Mocks = append(payload.Mocks, mocks...)
	return payload, nil
}

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

func main() {
	wd, _ := os.Getwd()

	payloads, err := walk(wd)
	if err != nil {
		panic(err.Error())
	}

	contents, err := yaml.Marshal(struct {
		Mocks []payload `yaml:"mocks"`
	}{
		Mocks: payloads,
	})
	if err != nil {
		panic(err.Error())
	}

	if err := os.WriteFile("mockgen.yaml", contents, os.ModePerm); err != nil {
		panic(err.Error())
	}
}

func walk(dir string) ([]payload, error) {
	infos, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var allPayloads []payload
	for _, info := range infos {
		path := filepath.Join(dir, info.Name())

		if info.IsDir() {
			payloads, err := walk(path)
			if err != nil {
				return nil, err
			}

			allPayloads = append(allPayloads, payloads...)
		} else if filepath.Ext(info.Name()) == ".go" {
			payloads, err := clean(path)
			if err != nil {
				return nil, err
			}

			if len(payloads) > 0 {
				if err := os.Remove(path); err != nil {
					return nil, err
				}

				allPayloads = append(allPayloads, payloads...)
			}
		}
	}

	return allPayloads, nil
}

var pattern = regexp.MustCompile(`^//go:generate (?:[\./]+)dev/mockgen.sh (.+)$`)

func clean(path string) ([]payload, error) {
	dir, _ := filepath.Rel("/Users/efritz/dev/sourcegraph/sourcegraph/", filepath.Dir(path))

	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var payloads []payload
	lines := bytes.Split(contents, []byte{'\n'})
	for _, line := range lines {
		if match := pattern.FindSubmatch(line); len(match) > 0 {
			payload, err := parseCommand(dir, string(match[1]))
			if err != nil {
				return nil, err
			}

			payloads = append(payloads, payload)
		}
	}

	return payloads, nil
}

type payload struct {
	Path        []string `yaml:"path,omitempty"`
	Interfaces  []string `yaml:"interfaces,omitempty"`
	Dirname     string   `yaml:"dirname,omitempty"`
	Filename    string   `yaml:"filename,omitempty"`
	PackageName string   `yaml:"package,omitempty"`
	Prefix      string   `yaml:"prefix,omitempty"`
	Force       bool     `yaml:"force,omitempty"`
}

func parseCommand(dirPath, match string) (payload, error) {
	var (
		path        string
		interfaces  []string
		dirname     string
		filename    string
		packageName string
		prefix      string
	)

	parts := strings.Split(match, " ")

	filtered := parts[:0]
	for _, part := range parts {
		if part != "" {
			filtered = append(filtered, part)
		}
	}
	parts = filtered

	path, parts = parts[0], parts[1:]

	for len(parts) > 0 {
		switch parts[0] {
		case "-i":
			interfaces = append(interfaces, parts[1])
			parts = parts[2:]
		case "-d":
			dirname, parts = filepath.Join(dirPath, parts[1]), parts[2:]
		case "-o":
			filename, parts = filepath.Join(dirPath, parts[1]), parts[2:]
		case "-p":
			packageName, parts = parts[1], parts[2:]
		case "--prefix":
			prefix, parts = parts[1], parts[2:]

		default:
			return payload{}, fmt.Errorf("Unknown flag %q", parts[0])
		}
	}

	return payload{
		Path:        []string{path},
		Interfaces:  interfaces,
		Dirname:     dirname,
		Filename:    filename,
		PackageName: packageName,
		Prefix:      prefix,
		Force:       true,
	}, nil
}

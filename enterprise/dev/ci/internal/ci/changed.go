package ci

import (
	"path/filepath"
	"strings"
)

type ChangedFiles []string

// affectsDocs returns whether the ChangedFiles affects documentation.
func (c ChangedFiles) affectsDocs() bool {
	for _, p := range c {
		if strings.HasPrefix(p, "doc/") && p != "CHANGELOG.md" {
			return true
		}
	}
	return false
}

// affectsSg returns whether the ChangedFiles affects the ./dev/sg folder.
func (c ChangedFiles) affectsSg() bool {
	for _, p := range c {
		if strings.HasPrefix(p, "dev/sg/") {
			return true
		}
	}
	return false
}

// affectsGo returns whether the ChangedFiles affects go files.
func (c ChangedFiles) affectsGo() bool {
	for _, p := range c {
		if strings.HasSuffix(p, ".go") || p == "go.sum" || p == "go.mod" {
			return true
		}
	}
	return false
}

// affectsDockerfiles whether the ChangedFiles affects Dockerfiles.
func (c ChangedFiles) affectsDockerfiles() bool {
	for _, p := range c {
		if strings.HasPrefix(p, "Dockerfile") || strings.HasSuffix(p, "Dockerfile") {
			return true
		}
	}
	return false
}

// Check if files that affect client code were changed. Used to detect if we need to run Puppeteer or Chromatic tests.
func (c ChangedFiles) affectsClient() bool {
	for _, p := range c {
		if !strings.HasSuffix(p, ".md") && (strings.HasPrefix(p, "client/") || isAllowedRootFile(p)) {
			return true
		}
	}
	return false
}

// Changes in the files below will be ignored by the Storybook workflow.
var ignoredRootFiles = []string{
	"jest.config.base.js",
	"graphql-schema-linter.config.js",
	"libsqlite3-pcre.dylib",
	".mocharc.js",
	"go.mod",
	"LICENSE",
	"renovate.json",
	"jest.config.js",
	"LICENSE.apache",
	".stylelintrc.json",
	".percy.yml",
	".tool-versions",
	"go.sum",
	".golangci.yml",
	".stylelintignore",
	".gitmodules",
	".prettierignore",
	".editorconfig",
	"prettier.config.js",
	".dockerignore",
	"doc.go",
	".gitignore",
	".gitattributes",
	".eslintrc.js",
	"sg.config.yaml",
	".eslintignore",
	".mailmap",
	"LICENSE.enterprise",
	"CODENOTIFY",
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func isAllowedRootFile(p string) bool {
	return filepath.Dir(p) == "." && !contains(ignoredRootFiles, p)
}

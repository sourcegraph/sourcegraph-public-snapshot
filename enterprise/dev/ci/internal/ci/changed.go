package ci

import (
	"path/filepath"
	"strings"
)

type ChangedFiles []string

// onlyDocs returns whether the ChangedFiles are only documentation.
func (c ChangedFiles) onlyDocs() bool {
	for _, p := range c {
		if !strings.HasPrefix(p, "doc/") && p != "CHANGELOG.md" {
			return false
		}
	}
	return true
}

// onlyConfig returns whether the ChangedFiles are only config file changes that don't need to run CI.
func (c ChangedFiles) onlyConfig() bool {
	for _, p := range c {
		if !strings.HasPrefix(p, ".github/") {
			return false
		}
	}
	return true
}

// onlySg returns whether the ChangedFiles are only in the ./dev/sg folder.
func (c ChangedFiles) onlySg() bool {
	for _, p := range c {
		if !strings.HasPrefix(p, "dev/sg/") {
			return false
		}
	}
	return true
}

// onlyGo returns whether the ChangedFiles are only go files.
func (c ChangedFiles) onlyGo() bool {
	for _, p := range c {
		if !strings.HasSuffix(p, ".go") && p != "go.sum" && p != "go.mod" {
			return false
		}
	}
	return true
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

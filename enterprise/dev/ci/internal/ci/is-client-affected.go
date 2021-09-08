package ci

import (
	"path/filepath"
	"strings"
)

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

// Check if files that affect client code were changed. Used to detect if we need to run Puppeteer or Chromatic tests.
func (c Config) isClientAffected() bool {
	for _, p := range c.changedFiles {
		if !strings.HasSuffix(p, ".md") && (strings.HasPrefix(p, "client/") || isAllowedRootFile(p)) {
			return true
		}
	}
	return false
}

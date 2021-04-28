package ci

import (
	"path/filepath"
	"strings"
)

// Changes in the files below will be ignored by the Storybook workflow.
var ignoredRootFiles []string = []string{
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
	".graphqlrc.yml",
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

func isNotIgnoredRootFile(p string) bool {
	return filepath.Dir(p) == "." && !contains(ignoredRootFiles, p)
}

// Run Storybook workflow only if related files were changed.
func (c Config) isStorybookAffected() bool {
	for _, p := range c.changedFiles {
		if !strings.HasSuffix(p, ".md") && (strings.HasPrefix(p, "client/") || isNotIgnoredRootFile(p)) {
			return true
		}
	}
	return false
}

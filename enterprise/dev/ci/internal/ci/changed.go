package ci

import (
	"path/filepath"
	"strings"
)

type ChangedFiles []string

// isDocsOnly returns whether the ChangedFiles are only documentation.
func (c ChangedFiles) isDocsOnly() bool {
	for _, p := range c {
		if !strings.HasPrefix(p, "doc/") && p != "CHANGELOG.md" {
			return false
		}
	}
	return true
}

// isSgOnly returns whether the ChangedFiles are only in the ./dev/sg folder.
func (c ChangedFiles) isSgOnly() bool {
	for _, p := range c {
		if !strings.HasPrefix(p, "dev/sg/") {
			return false
		}
	}
	return true
}

// isGoOnly returns whether the ChangedFiles are only go files.
func (c ChangedFiles) isGoOnly() bool {
	for _, p := range c {
		if !strings.HasSuffix(p, ".go") && p != "go.sum" && p != "go.mod" {
			return false
		}
	}
	return true
}

// Run Storybook workflow only if related files were changed.
func (c ChangedFiles) isStorybookAffected() bool {
	for _, p := range c {
		if !strings.HasSuffix(p, ".md") && (strings.HasPrefix(p, "client/") || isStorybookAllowedRootFile(p)) {
			return true
		}
	}
	return false
}

// Changes in the files below will be ignored by the Storybook workflow.
var storybookIgnoredRootFiles = []string{
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

func isStorybookAllowedRootFile(p string) bool {
	return filepath.Dir(p) == "." && !contains(storybookIgnoredRootFiles, p)
}

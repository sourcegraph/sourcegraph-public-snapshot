package embed

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExcludingFilePaths(t *testing.T) {
	files := []string{
		"file.sql",
		"root/file.yaml",
		"client/web/struct.json",
		"vendor/vendor.txt",
		"cool.go",
		"node_modules/a.go",
		"Dockerfile",
		"README.md",
		"vendor/README.md",
		"LICENSE.txt",
		"nested/vendor/file.py",
		".prettierignore",
		"client/web/.gitattributes",
		"no_ignore",
		"data/names.csv",
	}

	expectedFiles := []string{"cool.go", "Dockerfile", "README.md", "LICENSE.txt", "no_ignore"}
	gotFiles := []string{}

	excludedGlobPatterns := GetDefaultExcludedFilePathPatterns()
	for _, file := range files {
		if !isExcludedFilePath(file, excludedGlobPatterns) {
			gotFiles = append(gotFiles, file)
		}
	}

	require.Equal(t, expectedFiles, gotFiles)
}

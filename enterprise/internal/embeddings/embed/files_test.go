package embed

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExcludingFilePaths(t *testing.T) {
	files := []string{
		"file.sql",
		"root/file.yaml",
		"client/web/struct.JSON",
		"vendor/vendor.txt",
		"cool.go",
		"node_modules/a.go",
		"Dockerfile",
		"README.md",
		"vendor/README.md",
		"LICENSE.txt",
		"nested/vendor/file.py",
	}

	expectedFiles := []string{"cool.go", "Dockerfile", "README.md", "LICENSE.txt"}
	gotFiles := []string{}

	for _, file := range files {
		if !isExcludedFilePath(file) {
			gotFiles = append(gotFiles, file)
		}
	}

	require.Equal(t, expectedFiles, gotFiles)
}

package main

import (
	"os"
	"path/filepath"

	"github.com/sourcegraph/log"
)

func readFileOrFatal(path string, logger log.Logger) string {
	content, err := os.ReadFile(filepath.Join("tester/testdata", path))
	if err != nil {
		logger.Fatal("failed to read test fixture file", log.Error(err))
	}
	return string(content)
}

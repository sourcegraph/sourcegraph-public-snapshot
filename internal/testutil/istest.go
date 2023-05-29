package testutil

import (
	"os"
	"path/filepath"
	"strings"
)

var IsTest = func() bool {
	path, _ := os.Executable()
	return strings.HasSuffix(filepath.Base(path), "_test") || // Test binary build by Bazel
		filepath.Ext(path) == ".test" ||
		strings.Contains(path, "/T/___") || // Test path used by GoLand
		filepath.Base(path) == "__debug_bin" // Debug binary used by VSCode
}()

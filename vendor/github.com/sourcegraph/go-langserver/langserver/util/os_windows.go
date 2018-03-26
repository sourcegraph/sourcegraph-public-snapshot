// +build windows

package util

import (
	"path/filepath"
	"strings"
)

func virtualPath(path string) string {
	// Windows implementation converts path to slashes and prefixes it with slash
	path = filepath.ToSlash(path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	// Also, on Windows paths are case-insensitive
	return strings.ToLower(path)
}

func realPath(path string) string {
	// Windows implementation converts path to back slashes and removes a prefix slash
	path = filepath.FromSlash(strings.TrimPrefix(path, "/"))
	// Also, on Windows paths are case-insensitive
	return strings.ToLower(path)
}

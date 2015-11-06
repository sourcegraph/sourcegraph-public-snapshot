// +build windows

package util

import (
	"path/filepath"
	"strings"
)

func safeCommandName(command string) string {
	return strings.Replace(filepath.ToSlash(command), ":", "\\:", 1)
}

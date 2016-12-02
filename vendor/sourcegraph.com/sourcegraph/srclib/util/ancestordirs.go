package util

import (
	"path/filepath"
)

// AncestorDirs returns a list of p's ancestor
// directories (optionally including itself) excluding the root ("." or "/")).
func AncestorDirs(p string, self bool) []string {
	if p == "" {
		return nil
	}
	var dirs []string
	dir := filepath.Dir(p)
	for dir != "." && dir[len(dir)-1:] != string(filepath.Separator) {
		dirs = append([]string{dir}, dirs...)
		dir = filepath.Dir(dir)
	}
	if self {
		dirs = append(dirs, p)
	}
	return dirs
}

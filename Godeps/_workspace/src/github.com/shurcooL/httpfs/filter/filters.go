package filter

import (
	"os"
	"path/filepath"
)

// FilesWithExtensions returns a filter that ignores files (but not directories) that
// have any of the given extensions. For example:
//
// 	filter.FilesWithExtensions(".go", ".html")
//
// Would ignore both .go and .html files. It would not ignore any directories.
func FilesWithExtensions(exts ...string) Func {
	return func(path string, fi os.FileInfo) bool {
		if fi.IsDir() {
			return false
		}
		for _, ext := range exts {
			if filepath.Ext(path) == ext {
				return true
			}
		}
		return false
	}
}

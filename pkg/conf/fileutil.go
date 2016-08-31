package conf

import (
	"os"
	"path/filepath"

	_ "sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/env"
)

func init() {
	if sgpath := os.Getenv("SGPATH"); sgpath != "" {
		tmpdir := filepath.Join(sgpath, "tmp")
		if _, err := os.Lstat(tmpdir); os.IsNotExist(err) {
			os.MkdirAll(tmpdir, os.FileMode(0755))
		}
	}
}

// TempDir returns the tmp directory rooted at $SGPATH. This should be used
// for things like file-based caches.
func TempDir() string {
	if sgpath := os.Getenv("SGPATH"); sgpath != "" {
		return filepath.Join(sgpath, "tmp")
	}
	return os.TempDir()
}

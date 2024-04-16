package codeintel

import (
	"os"
	"path/filepath"
)

// SanitizeRoot removes redundant paths from the given root and replaces
// references to the root of th repo with an empty string.
func SanitizeRoot(root string) string {
	if root = filepath.Clean(root); root == "." || (root != "" && os.IsPathSeparator(root[0])) {
		return ""
	}
	return root
}

// Package usercontent contains a store for user-uploaded content.
package usercontent

import (
	"fmt"
	"os"
	"path/filepath"

	"sourcegraph.com/sourcegraph/rwvfs"
)

// Store for user uploaded content. Unset by default.
var Store rwvfs.FileSystem

// LocalStore creates or reuses a local disk-based store for user content
// that uses $SGPATH/usercontent directory for storage.
func LocalStore() (rwvfs.FileSystem, error) {
	dir := filepath.Join(os.Getenv("SGPATH"), "usercontent")
	err := os.Mkdir(dir, 0755)
	if err != nil && !os.IsExist(err) {
		return nil, fmt.Errorf("creating directory %q failed: %v", dir, err)
	}
	return rwvfs.OS(dir), nil
}

package inference

import (
	"path/filepath"
)

// containsSegment returns true if the given path contains the given segment.
func containsSegment(path, segment string) bool {
	if path == "" {
		return false
	}

	dir, file := filepath.Split(filepath.Clean(path))
	if file == segment {
		return true
	}

	return containsSegment(dir, segment)
}

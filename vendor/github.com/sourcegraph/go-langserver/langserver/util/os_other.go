// +build !windows

package util

func virtualPath(path string) string {
	// non-Windows implementation does nothing
	return path
}

func realPath(path string) string {
	// non-Windows implementation does nothing
	return path
}

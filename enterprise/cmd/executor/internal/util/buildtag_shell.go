//go:build shell
// +build shell

package util

// HasShellBuildTag returns true if the shell build tag is set.
func HasShellBuildTag() bool {
	return true
}

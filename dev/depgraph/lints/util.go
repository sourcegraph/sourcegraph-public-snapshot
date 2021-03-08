package lints

import "regexp"

var cmdPrefixPattern = regexp.MustCompile(`\bcmd/([^/]+)`)

// containingCommand returns the package root of the command the given path resides in,
// if any. This will return the same value for packages composing the same binary and
// different values for different binaries and shared code.
func containingCommand(path string) string {
	if match := cmdPrefixPattern.FindAllStringSubmatch(path, 1); len(match) > 0 {
		return match[0][1]
	}

	return ""
}

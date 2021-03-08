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

var publicLibraryPrefix = regexp.MustCompile("^(?:enterprise/)?lib/")

// isLibray returns true if the given path is publicly importable.
func isLibrary(path string) bool {
	return publicLibraryPrefix.MatchString(path)
}

// isMain returns true if the given package declares "main" in the given package name map.
func isMain(packageNames map[string][]string, pkg string) bool {
	for _, name := range packageNames[pkg] {
		if name == "main" {
			return true
		}
	}

	return false
}

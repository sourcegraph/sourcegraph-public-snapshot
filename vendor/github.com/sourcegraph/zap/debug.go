package zap

import "strings"

// CheckBranchName is a temporary debugging helper that panics if
// branch doesn't look like a branch name.
func CheckBranchName(branch BranchName) {
	if branch == "" {
		panic("suspicious branch name: (empty)")
	}
	if strings.HasPrefix(string(branch), "branch/") {
		panic("suspicious branch name: " + string(branch) + " (should not have a branch/ prefix)")
	}
	if strings.HasPrefix(string(branch), "head/") {
		panic("suspicious branch name: " + string(branch) + " (should not have a head/ prefix)")
	}
}

// CheckSymbolicRefName is a temporary debugging helper that panics if
// ref doesn't look like a symbolic ref name.
func CheckSymbolicRefName(ref string) {
	if ref == "" {
		panic("suspicious symbolic ref name: (empty)")
	}
}

// CheckRefName is a temporary debugging helper that panics if
// ref doesn't look like a ref name.
func CheckRefName(ref string) {
	if ref == "" {
		panic("suspicious ref name: (empty)")
	}
}

package zap

import "strings"

var tempStrictNameChecks = true

// CheckBranchName is a temporary debugging helper that panics if
// branch doesn't look like a branch name.
func CheckBranchName(branch BranchName) {
	if branch == "" {
		panic("suspicious branch name: (empty)")
	}
	if strings.HasPrefix(string(branch), "branch/") || strings.HasPrefix(string(branch), "remote/") {
		panic("suspicious branch name: " + string(branch) + " (should not have a branch/ or remote/ prefix)")
	}
	if strings.ToUpper(string(branch)) == string(branch) {
		panic("suspicious branch name: " + string(branch) + " (should not be ALL UPPERCASE)")
	}
	if strings.Contains(string(branch), "HEAD") {
		panic("suspicious branch name: " + string(branch) + " (should not contain HEAD)")
	}
}

// CheckSymbolicRefName is a temporary debugging helper that panics if
// ref doesn't look like a symbolic ref name.
func CheckSymbolicRefName(ref string) {
	if ref == "" {
		panic("suspicious symbolic ref name: (empty)")
	}
	if strings.HasPrefix(ref, "branch/") || strings.HasPrefix(ref, "remote/") || strings.ToUpper(ref) != ref {
		panic("suspicious symbolic ref name: " + ref + " (should not have a branch/ or remote/ prefix, should be ALL UPPERCASE)")
	}
}

// CheckRefName is a temporary debugging helper that panics if
// ref doesn't look like a ref name.
func CheckRefName(ref string) {
	if ref == "" {
		panic("suspicious ref name: (empty)")
	}
	if strings.ToUpper(ref) == ref && !strings.Contains(ref, "/") {
		return // valid symbolic ref
	}
	if !strings.HasPrefix(ref, "branch/") && !strings.HasPrefix(ref, "remote/") {
		panic("suspicious ref name: " + ref + " (should have a branch/ or remote/ prefix)")
	}
	if strings.Contains(ref, "branch/branch") || strings.Contains(ref, "remote/remote") || strings.Contains(ref, "branch/remote") || strings.Contains(ref, "remote/branch") {
		panic("suspicious ref name: " + ref)
	}
	if strings.HasPrefix(ref, "remote/") {
		if ref == "remote/*" {
			return // valid refspec
		}
		parts := strings.SplitN(ref, "/", 4)
		if len(parts) != 4 {
			panic("suspicious remote tracking ref name: " + ref + " (should have 4 or more path components)")
		}
		if parts[2] != "branch" {
			panic("suspicious remote tracking ref name: " + ref + " (should be of the form remote/$REMOTE/branch/$BRANCH)")
		}
	}
}

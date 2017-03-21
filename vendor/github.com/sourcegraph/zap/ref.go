package zap

import (
	"fmt"
	"strings"
)

// IsRemoteRef reports whether ref refers to a remote ref (starts with
// "remote/").
func IsRemoteRef(ref string) bool {
	CheckRefName(ref)
	return strings.HasPrefix(ref, "remote/")
}

// IsBranchRef reports whether ref refers to a branch (starts with
// "branch/").
func IsBranchRef(ref string) bool {
	CheckRefName(ref)
	return strings.HasPrefix(ref, "branch/")
}

// A BranchName is the name of a Zap branch. A branch named "b" is a
// ref named "branch/b". Values of type BranchName omit the "branch/"
// prefix.
type BranchName string

// Ref returns the ref name for b ("branch/" + b).
func (b BranchName) Ref() string {
	return "branch/" + string(b)
}

// RemoteTrackingRef returns the remote tracking ref name for b on
// remote ("remote/" + remote + "/branch/" + b).
func (b BranchName) RemoteTrackingRef(remote string) string {
	return "remote/" + remote + "/branch/" + string(b)
}

// BranchNameFromRef parses the branch name from a branch ref. If ref
// is not a branch ref, an error is returned.
func BranchNameFromRef(ref string) (BranchName, error) {
	if !strings.HasPrefix(ref, "branch/") {
		return "", fmt.Errorf("not a branch ref: %s", ref)
	}
	return BranchName(strings.TrimPrefix(ref, "branch/")), nil
}

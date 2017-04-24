package zap

import (
	"fmt"
	"strings"
)

// IsBranchRef reports whether ref refers to a branch (starts with
// "branch/").
func IsBranchRef(ref string) bool {
	return strings.HasPrefix(ref, "branch/") && ref != "branch/"
}

// IsHeadRef reports whether ref refers to a head (starts with
// "head/").
func IsHeadRef(ref string) bool {
	return strings.HasPrefix(ref, "head/") && ref != "head/"
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
	if !IsBranchRef(ref) {
		return "", fmt.Errorf("not a branch ref: %s", ref)
	}
	return BranchName(strings.TrimPrefix(ref, "branch/")), nil
}

// ClientIDFromHeadRef returns "c" from "head/c".
func ClientIDFromHeadRef(ref string) (string, error) {
	if !strings.HasPrefix(ref, "head/") || ref == "head/" {
		return "", fmt.Errorf("invalid head ref: %s", ref)
	}
	return strings.TrimPrefix(ref, "head/"), nil
}

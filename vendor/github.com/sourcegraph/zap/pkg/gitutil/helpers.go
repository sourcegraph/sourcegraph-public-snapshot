package gitutil

import (
	"fmt"
	"strings"
)

// CurrentBranch returns the name of the current branch (i.e., the
// value of the HEAD symbolic ref) in a Git repository.
func CurrentBranch(gitRepo interface {
	ReadSymbolicRef(string) (string, error)
}) (string, error) {
	v, err := gitRepo.ReadSymbolicRef("HEAD")
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(v, "refs/heads/") {
		return v, nil
	}
	return "", fmt.Errorf("invalid HEAD %q", v)
}

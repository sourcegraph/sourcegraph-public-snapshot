package sourcegraph

import (
	"errors"
	"strings"
)

// Bool is a helper routine that allocates a new bool value to store v
// and returns a pointer to it.
func Bool(v bool) *bool {
	p := new(bool)
	*p = v
	return p
}

// String is a helper routine that allocates a new string value to
// store v and returns a pointer to it.
func String(v string) *string {
	p := new(string)
	*p = v
	return p
}

// Int is a helper routine that allocates a new int value to store v
// and returns a pointer to it.
func Int(v int) *int {
	p := new(int)
	*p = v
	return p
}

// ParseRepoAndCommitID parses strings like "example.com/repo" and
// "example.com/repo@myrev".
func ParseRepoAndCommitID(repoAndCommitID string) (uri, commitID string) {
	if i := strings.Index(repoAndCommitID, "@"); i != -1 {
		return repoAndCommitID[:i], repoAndCommitID[i+1:]
	}
	return repoAndCommitID, ""
}

// Primary returns the primary email address in the list, or an error.
func (e *EmailAddrList) Primary() (*EmailAddr, error) {
	for _, email := range e.EmailAddrs {
		if email.Primary {
			return email, nil
		}
	}
	return nil, errors.New("no primary email address")
}

package xlang

import (
	"fmt"
	"strings"
	"sync"
)

type errorList struct {
	mu     sync.Mutex
	errors []error
}

// add adds err to the list of errors. It is safe to call it from
// concurrent goroutines.
func (e *errorList) add(err error) {
	e.mu.Lock()
	e.errors = append(e.errors, err)
	e.mu.Unlock()
}

// errors returns the list of errors as a single error. It is NOT safe
// to call from concurrent goroutines.
func (e *errorList) error() error {
	switch len(e.errors) {
	case 0:
		return nil
	case 1:
		return e.errors[0]
	default:
		return fmt.Errorf("%s [and %d more errors]", e.errors[0], len(e.errors)-1)
	}
}

// ConvertFileToGitURI takes a file URI like like `file://repo/path/file/path`
// and returns a git URI like `git://repo/path?rev#file/path`.
// This is useful because the client uses file URIs, but xlang
// currently requires git URIs.
func ConvertFileToGitURI(uri string, repo string, rev string) (string, error) {
	if !strings.HasPrefix(uri, "file://") {
		return "", fmt.Errorf("expected file:// URI, got %s", uri)
	}
	if !strings.HasPrefix(strings.TrimPrefix(uri, "file://"), repo) {
		return "", fmt.Errorf("unexpected URI for repo %s: %s", repo, uri)
	}
	if len(rev) != 40 {
		return "", fmt.Errorf("invalid git revision: %s", rev)
	}
	uriParts := strings.Split(uri, "/")
	repoParts := strings.Split(repo, "/")
	filePath := strings.Join(uriParts[:(2+len(repoParts))], "/")
	repoURI := "git://" + repo + "?" + rev
	if filePath != "" {
		return repoURI + "#" + filePath, nil
	}
	return repoURI, nil
}

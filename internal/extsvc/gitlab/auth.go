package gitlab

import (
	"fmt"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
)

// sudoableToken represents a personal access token with an optional sudo scope.
type sudoableToken struct {
	Token string
	Sudo  string
}

var _ auth.Authenticator = &sudoableToken{}

func (pat *sudoableToken) Authenticate(req *http.Request) error {
	req.Header.Set("Private-Token", pat.Token)

	if pat.Sudo != "" {
		req.Header.Set("Sudo", pat.Sudo)
	}

	return nil
}

func (pat *sudoableToken) Hash() string {
	return fmt.Sprintf("pat::sudo:%s::%s", pat.Sudo, pat.Token)
}

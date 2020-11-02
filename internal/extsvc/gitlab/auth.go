package gitlab

import (
	"fmt"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
)

// SudoableToken represents a personal access token with an optional sudo scope.
type SudoableToken struct {
	Token string
	Sudo  string
}

var _ auth.Authenticator = &SudoableToken{}

func (pat *SudoableToken) Authenticate(req *http.Request) error {
	req.Header.Set("Private-Token", pat.Token)

	if pat.Sudo != "" {
		req.Header.Set("Sudo", pat.Sudo)
	}

	return nil
}

func (pat *SudoableToken) Hash() string {
	return fmt.Sprintf("pat::sudoku:%s::%s", pat.Sudo, pat.Token)
}

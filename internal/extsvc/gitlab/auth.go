package gitlab

import (
	"fmt"
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
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

// RequestedOAuthScopes returns the list of OAuth scopes given the default API
// scope and any extra scopes.
func RequestedOAuthScopes(defaultAPIScope string, extraScopes []string) []string {
	scopes := []string{"read_user"}
	if defaultAPIScope == "" {
		defaultAPIScope = "api"
	}
	if envvar.SourcegraphDotComMode() {
		// By default, request `read_api`. User's who are allowed to add private code
		// will request full `api` access via extraScopes.
		scopes = append(scopes, "read_api")
	} else {
		// For customer instances we default to api scope so that they can clone private
		// repos but in they can optionally override this in config.
		scopes = append(scopes, defaultAPIScope)
	}
	// Append extra scopes and ensure there are no duplicates
	for _, s := range extraScopes {
		var found bool
		for _, inner := range scopes {
			if inner == s {
				found = true
				break
			}
		}

		if !found {
			scopes = append(scopes, s)
		}
	}

	return scopes
}

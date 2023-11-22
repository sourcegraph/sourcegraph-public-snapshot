package gitlab

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
)

// Metrics here exported as they are needed from two different packages

var TokenRefreshCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "src_repoupdater_gitlab_token_refresh_count",
	Help: "Counts the number of times we refresh a GitLab OAuth token",
}, []string{"source", "success"})

var TokenMissingRefreshCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "src_repoupdater_gitlab_token_missing_refresh_count",
	Help: "Counts the number of times we see a token without a refresh token",
})

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
func RequestedOAuthScopes(defaultAPIScope string) []string {
	scopes := []string{"read_user"}
	if defaultAPIScope == "" {
		scopes = append(scopes, "api")
	} else {
		scopes = append(scopes, defaultAPIScope)
	}

	return scopes
}

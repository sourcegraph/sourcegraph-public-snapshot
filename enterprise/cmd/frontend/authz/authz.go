package authz

import (
	iauthz "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz"
	"github.com/sourcegraph/sourcegraph/schema"
)

// ValidateGitLab performs static validation of the given GitLab authorization config using the provided instance
// url, token and auth providers.
func ValidateGitLab(a *schema.GitLabAuthorization, instanceURL, token string, ps []schema.AuthProviders) error {
	return iauthz.ValidateGitLabAuthz(a, instanceURL, token, ps)
}

// ValidateGitHub performs static validation of the given GitHub authorization config
// using the provided instance url and token.
func ValidateGitHub(a *schema.GitHubAuthorization, instanceURL, token string) error {
	return iauthz.ValidateGitHubAuthz(a, instanceURL, token)
}

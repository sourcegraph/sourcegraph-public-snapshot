package conf

import (
	"net/url"

	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

// EmailVerificationRequired returns whether users must verify an email address before they
// can perform most actions on this site.
//
// It's false for sites that do not have an email sending API key set up.
func EmailVerificationRequired() bool {
	return Get().EmailSmtp != nil
}

// CanSendEmail returns whether the site can send emails (e.g., to reset a password or
// invite a user to an org).
//
// It's false for sites that do not have an email sending API key set up.
func CanSendEmail() bool {
	return Get().EmailSmtp != nil
}

// FirstGitHubDotComConnectionWithToken returns the first GitHub connection
// config for github.com that contains a token, or nil if there is none.
func FirstGitHubDotComConnectionWithToken() *schema.GitHubConnection {
	for _, c := range Get().Github {
		if c.Token == "" {
			continue
		}
		u, _ := url.Parse(c.Url)
		if u != nil && u.Hostname() == "github.com" {
			return &c
		}
	}
	return nil
}

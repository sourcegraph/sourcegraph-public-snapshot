package github

import (
	"strings"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"

	"github.com/sourcegraph/go-github/github"
)

// GetGitHubUserEmailAddresses attempts to determine the user's email
// addresses from the GitHub API. If nonPublicEmails, it also looks up
// the list of the user's emails (and doesn't just look at the user's
// public email).
func GetGitHubUserEmailAddresses(ghuser *github.User, gh *github.Client, nonPublicEmails bool) ([]*sourcegraph.EmailAddr, error) {
	if ghuser == nil {
		panic("ghuser == nil")
	}
	if gh == nil {
		panic("gh == nil")
	}

	// For users whose OAuth tokens we have, prefer the list emails API.
	if nonPublicEmails {
		authenticatedUser, _, err := gh.Users.Get("")
		if err == nil && authenticatedUser.Login != nil && ghuser.Login != nil && *(authenticatedUser.Login) == *(ghuser.Login) {
			emails, _, err := gh.Users.ListEmails(nil)
			if err != nil {
				return nil, err
			}

			var addrs []*sourcegraph.EmailAddr
			for _, ghemail := range emails {
				email := strings.ToLower(*ghemail.Email)
				if email != "" {
					addrs = append(addrs, &sourcegraph.EmailAddr{
						Email:    email,
						Verified: *ghemail.Verified,
						Primary:  *ghemail.Primary,
					})
				}
			}
			return addrs, nil
		}
	}

	// Otherwise, just use their public email.
	if ghuser.Email != nil && *ghuser.Email != "" {
		return []*sourcegraph.EmailAddr{
			{
				Email:    *ghuser.Email,
				Verified: false, // GitHub public email is not verified
				Primary:  false, // seeing Primary email requires user:email scope
			},
		}, nil
	}

	return nil, nil
}

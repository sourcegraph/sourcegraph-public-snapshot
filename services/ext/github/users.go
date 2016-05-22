package github

import (
	"github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func userFromGitHub(ghuser *github.User) *sourcegraph.User {
	var u sourcegraph.User
	u.UID = int32(*ghuser.ID)
	u.Login = *ghuser.Login
	if ghuser.Name != nil {
		u.Name = *ghuser.Name
	}
	if ghuser.AvatarURL != nil {
		u.AvatarURL = *ghuser.AvatarURL
	}
	if ghuser.Location != nil {
		u.Location = *ghuser.Location
	}
	if ghuser.Company != nil {
		u.Company = *ghuser.Company
	}
	if ghuser.Blog != nil {
		u.HomepageURL = *ghuser.Blog
	}
	if ghuser.Type != nil {
		u.IsOrganization = *ghuser.Type == "Organization"
	}
	return &u
}

package graphqlbackend

import "github.com/sourcegraph/go-github/github"

type installationResolver struct {
	installation *github.Installation
}

func (i *installationResolver) Login() string {
	if i.installation.Account.Login == nil {
		return ""
	}
	return *i.installation.Account.Login
}

func (i *installationResolver) GitHubID() int32 {
	if i.installation.Account.ID == nil {
		return 0
	}
	return int32(*i.installation.Account.ID)
}

func (i *installationResolver) InstallID() int32 {
	if i.installation.ID == nil {
		return 0
	}
	return int32(*i.installation.ID)
}

func (i *installationResolver) Type() string {
	if i.installation.Account.Type == nil {
		return ""
	}
	return *i.installation.Account.Type
}

func (i *installationResolver) AvatarURL() string {
	if i.installation.Account.AvatarURL == nil {
		return ""
	}
	return *i.installation.Account.AvatarURL
}

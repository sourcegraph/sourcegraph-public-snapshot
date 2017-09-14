package tracking

import (
	"context"

	"github.com/sourcegraph/go-github/github"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	extgithub "sourcegraph.com/sourcegraph/sourcegraph/pkg/github"
)

// ListAllOrgs is a convenience wrapper around listOrgsPage (since GitHub's API is paginated), returning
// a list of all of the current user's GitHub organizations
//
// This method may return an error and a partial list of organizations
func listAllOrgs(ctx context.Context, op *sourcegraph.ListOptions) (res *sourcegraph.OrgsList, err error) {
	var orgs []*sourcegraph.GitHubOrg
	if !extgithub.HasAuthedUser(ctx) {
		return &sourcegraph.OrgsList{}, nil
	}
	installs, err := extgithub.ListAllAccessibleInstallations(ctx)
	if err != nil {
		return nil, err
	}
	for _, ins := range installs {
		if *ins.Account.Type == "Organization" {
			orgs = append(orgs, toOrgFromAccount(ins.Account))
		}
	}
	return &sourcegraph.OrgsList{Orgs: orgs}, nil
}

func toOrgFromAccount(user *github.User) *sourcegraph.GitHubOrg {
	strv := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}

	intv := func(i *int) int32 {
		if i == nil {
			return 0
		}
		return int32(*i)
	}

	org := sourcegraph.GitHubOrg{
		Login:         *user.Login,
		ID:            int32(*user.ID),
		AvatarURL:     strv(user.AvatarURL),
		Name:          strv(user.Name),
		Blog:          strv(user.Blog),
		Location:      strv(user.Location),
		Email:         strv(user.Email),
		Description:   "",
		Collaborators: intv(user.Collaborators),
	}

	return &org
}

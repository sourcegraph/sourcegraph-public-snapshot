package github

import (
	"github.com/sourcegraph/go-github/github"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

type Orgs struct{}

func (s *Orgs) List(ctx context.Context, user sourcegraph.UserSpec, opt *sourcegraph.ListOptions) ([]*sourcegraph.Org, error) {
	ghOrgs, _, err := client(ctx).orgs.List(user.Login, &github.ListOptions{
		PerPage: opt.PerPageOrDefault(), Page: opt.PageOrDefault(),
	})
	if err != nil {
		return nil, err
	}
	orgs := make([]*sourcegraph.Org, len(ghOrgs))
	for i, ghOrg := range ghOrgs {
		orgs[i] = &sourcegraph.Org{User: sourcegraph.User{
			Login: *ghOrg.Login,
		}}
	}
	return orgs, nil
}

func (s *Orgs) ListMembers(ctx context.Context, org sourcegraph.OrgSpec, opt *sourcegraph.OrgListMembersOptions) ([]*sourcegraph.User, error) {
	if org.Org == "" {
		panic("org.Org is empty")
	}

	if opt == nil {
		opt = &sourcegraph.OrgListMembersOptions{}
	}

	ghmembers, _, err := client(ctx).orgs.ListMembers(org.Org,
		&github.ListMembersOptions{
			ListOptions: github.ListOptions{PerPage: opt.PerPageOrDefault(), Page: opt.PageOrDefault()},
		},
	)
	if err != nil {
		return nil, err
	}

	members := make([]*sourcegraph.User, len(ghmembers))
	for i, ghmember := range ghmembers {
		members[i] = userFromGitHub(&ghmember)
	}

	return members, nil
}

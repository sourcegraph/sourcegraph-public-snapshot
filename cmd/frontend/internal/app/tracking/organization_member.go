package tracking

import (
	"context"
	"fmt"

	"github.com/sourcegraph/go-github/github"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	extgithub "sourcegraph.com/sourcegraph/sourcegraph/pkg/github"
)

// listOrgMembersPage returns a single page of an organization's members
func listOrgMembersPage(ctx context.Context, orgLogin string, opt *sourcegraph.ListOptions) (res []*github.User, err error) {
	if !extgithub.HasAuthedUser(ctx) {
		return []*github.User{}, nil
	}

	optGh := &github.ListMembersOptions{
		ListOptions: github.ListOptions{
			Page:    int(opt.Page),
			PerPage: int(opt.PerPage),
		},
	}

	cl := extgithub.Client(ctx)
	installs, _, err := cl.Users.ListInstallations(ctx, nil)
	if err != nil {
		return nil, err
	}
	for _, ins := range installs {
		if *ins.Account.Login == orgLogin {
			insClient, err := extgithub.InstallationClient(ctx, *ins.ID)
			if err != nil {
				return nil, err
			}
			members, _, err := insClient.Organizations.ListMembers(ctx, orgLogin, optGh)
			if err != nil {
				return nil, err
			}
			return members, nil
		}
	}
	return nil, fmt.Errorf("github org %s not found for current user", orgLogin)
}

// listAllOrgMembers is a convenience wrapper around listOrgMembersPage (since GitHub's API is paginated), returning
// a list of all of the specified org's GitHub members
//
// This method may return an error and a partial list of organization members
func listAllOrgMembers(ctx context.Context, orgLogin string, opt *sourcegraph.ListOptions) (res []*github.User, err error) {
	// Get a maximum of 1,000 members per org
	const perPage = 100
	const maxPage = 10
	opts := *opt
	opts.PerPage = perPage

	var allMembers []*github.User
	for page := 1; page <= maxPage; page++ {
		opts.Page = int32(page)
		members, err := listOrgMembersPage(ctx, orgLogin, &opts)
		if err != nil {
			// If an error occurs, return that error, as well as a list of all members
			// collected so far
			return allMembers, err
		}
		allMembers = append(allMembers, members...)
		if len(members) < perPage {
			break
		}
	}
	return allMembers, nil
}

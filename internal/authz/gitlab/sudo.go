package gitlab

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
)

// FetchUserPerms returns a list of project IDs (on code host) that the given account
// has read access on the code host. The project ID has the same value as it would be
// used as api.ExternalRepoSpec.ID. The returned list only includes private project IDs.
//
// This method may return partial but valid results in case of error, and it is up to
// callers to decide whether to discard.
//
// API docs: https://docs.gitlab.com/ee/api/projects.html#list-all-projects
func (p *SudoProvider) FetchUserPerms(ctx context.Context, account *extsvc.Account) ([]extsvc.RepoID, error) {
	if account == nil {
		return nil, errors.New("no account provided")
	} else if !extsvc.IsHostOfAccount(p.codeHost, account) {
		return nil, fmt.Errorf("not a code host of the account: want %q but have %q",
			account.AccountSpec.ServiceID, p.codeHost.ServiceID)
	}

	user, _, err := gitlab.GetExternalAccountData(&account.AccountData)
	if err != nil {
		return nil, errors.Wrap(err, "get external account data")
	}

	client := p.clientProvider.GetPATClient(p.sudoToken, strconv.Itoa(int(user.ID)))
	return listProjects(ctx, client)
}

// listProjects is a helper function to request for all private projects that are accessible
// (access level: 20 => Reporter access) by the authenticated or impersonated user in the client.
// It may return partial but valid results in case of error, and it is up to callers to decide
// whether to discard.
func listProjects(ctx context.Context, client *gitlab.Client) ([]extsvc.RepoID, error) {
	q := make(url.Values)
	q.Add("visibility", "private")  // This method is meant to return only private projects
	q.Add("min_access_level", "20") // 20 => Reporter access (i.e. have access to project code)
	q.Add("per_page", "100")        // 100 is the maximum page size

	// The next URL to request for projects, and it is reused in the succeeding for loop.
	nextURL := "projects?" + q.Encode()

	// 100 matches the maximum page size, thus a good default to avoid multiple allocations
	// when appending the first 100 results to the slice.
	projectIDs := make([]extsvc.RepoID, 0, 100)
	for {
		projects, next, err := client.ListProjects(ctx, nextURL)
		if err != nil {
			return projectIDs, err
		}

		for _, p := range projects {
			projectIDs = append(projectIDs, extsvc.RepoID(strconv.Itoa(p.ID)))
		}

		if next == nil {
			break
		}
		nextURL = *next
	}

	return projectIDs, nil
}

// FetchRepoPerms returns a list of user IDs (on code host) who have read access to
// the given project on the code host. The user ID has the same value as it would
// be used as extsvc.Account.AccountID. The returned list includes both direct access
// and inherited from the group membership.
//
// This method may return partial but valid results in case of error, and it is up to
// callers to decide whether to discard.
//
// API docs: https://docs.gitlab.com/ee/api/members.html#list-all-members-of-a-group-or-project-including-inherited-members
func (p *SudoProvider) FetchRepoPerms(ctx context.Context, repo *extsvc.Repository) ([]extsvc.AccountID, error) {
	if repo == nil {
		return nil, errors.New("no repository provided")
	} else if !extsvc.IsHostOfRepo(p.codeHost, &repo.ExternalRepoSpec) {
		return nil, fmt.Errorf("not a code host of the repository: want %q but have %q",
			repo.ServiceID, p.codeHost.ServiceID)
	}

	client := p.clientProvider.GetPATClient(p.sudoToken, "")
	return listMembers(ctx, client, repo.ID)
}

// listMembers is a helper function to request for all users who has read access
// (access level: 20 => Reporter access) to given project on the code host, including
// both direct access and inherited from the group membership. It may return partial
// but valid results in case of error, and it is up to callers to decide whether to
// discard.
func listMembers(ctx context.Context, client *gitlab.Client, repoID string) ([]extsvc.AccountID, error) {
	q := make(url.Values)
	q.Add("per_page", "100") // 100 is the maximum page size

	// The next URL to request for members, and it is reused in the succeeding for loop.
	nextURL := fmt.Sprintf("projects/%s/members/all?%s", repoID, q.Encode())

	// 100 matches the maximum page size, thus a good default to avoid multiple allocations
	// when appending the first 100 results to the slice.
	userIDs := make([]extsvc.AccountID, 0, 100)

	for {
		members, next, err := client.ListMembers(ctx, nextURL)
		if err != nil {
			return userIDs, err
		}

		for _, m := range members {
			// Members with access level 20 (i.e. Reporter) has access to project code.
			if m.AccessLevel < 20 {
				continue
			}

			userIDs = append(userIDs, extsvc.AccountID(strconv.Itoa(int(m.ID))))
		}

		if next == nil {
			break
		}
		nextURL = *next
	}

	return userIDs, nil
}

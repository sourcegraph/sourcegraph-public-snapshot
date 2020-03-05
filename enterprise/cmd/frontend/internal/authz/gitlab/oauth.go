package gitlab

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/api"
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
func (p *OAuthAuthzProvider) FetchUserPerms(ctx context.Context, account *extsvc.ExternalAccount) ([]string, error) {
	if account == nil {
		return nil, errors.New("no account provided")
	} else if account.ServiceType != p.codeHost.ServiceType || account.ServiceID != p.codeHost.ServiceID {
		return nil, fmt.Errorf("service mismatch: want %q - %q but the account has %q - %q",
			p.codeHost.ServiceType, p.codeHost.ServiceID, account.ServiceType, account.ServiceID)
	}

	_, tok, err := gitlab.GetExternalAccountData(&account.ExternalAccountData)
	if err != nil {
		return nil, errors.Wrap(err, "get external account data")
	}
	token := tok.AccessToken

	q := make(url.Values)
	q.Add("visibility", "private")  // This method is meant to return only private projects
	q.Add("min_access_level", "20") // 20 => Reporter access (i.e. have access to project code)
	q.Add("per_page", "100")        // 100 is the maximum page size

	// The next URL to request for projects, and it is reused in the succeeding for loop.
	nextURL := "projects?" + q.Encode()

	// 100 matches the maximum page size, thus a good default to avoid multiple allocations
	// when appending the first 100 results to the slice.
	projectIDs := make([]string, 0, 100)

	client := p.clientProvider.GetOAuthClient(token)
	for {
		projects, next, err := client.ListProjects(ctx, nextURL)
		if err != nil {
			return projectIDs, err
		}

		for i := range projects {
			projectIDs = append(projectIDs, strconv.Itoa(projects[i].ID))
		}

		if next == nil {
			break
		}
		nextURL = *next
	}

	return projectIDs, nil
}

// FetchRepoPerms is a stub implementation for OAuth authz provider because the API
// endpoint we use requires admin in order to get complete results.
func (p *OAuthAuthzProvider) FetchRepoPerms(context.Context, *api.ExternalRepoSpec) ([]string, error) {
	return []string{}, nil
}

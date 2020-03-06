package gitlab

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
)

// FetchUserPerms returns a list of private project IDs (on code host) that the given account
// has read access to. The project ID has the same value as it would be
// used as api.ExternalRepoSpec.ID. The returned list only includes private project IDs.
//
// This method may return partial but valid results in case of error, and it is up to
// callers to decide whether to discard.
//
// API docs: https://docs.gitlab.com/ee/api/projects.html#list-all-projects
func (p *OAuthAuthzProvider) FetchUserPerms(ctx context.Context, account *extsvc.ExternalAccount) ([]extsvc.ExternalRepoID, error) {
	if account == nil {
		return nil, errors.New("no account provided")
	} else if !extsvc.IsHostOfAccount(p.codeHost, account) {
		return nil, fmt.Errorf("not a code host of the account: want %+v but have %+v", account, p.codeHost)
	}

	_, tok, err := gitlab.GetExternalAccountData(&account.ExternalAccountData)
	if err != nil {
		return nil, errors.Wrap(err, "get external account data")
	}

	client := p.clientProvider.GetOAuthClient(tok.AccessToken)
	return listProjects(ctx, client)
}

// FetchRepoPerms is a stub implementation for OAuth authz provider because the API
// endpoint we use requires admin in order to get complete results.
func (p *OAuthAuthzProvider) FetchRepoPerms(context.Context, *api.ExternalRepoSpec) ([]string, error) {
	return []string{}, nil
}

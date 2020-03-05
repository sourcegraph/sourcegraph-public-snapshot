package gitlab

import (
	"context"
	"fmt"

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

	client := p.clientProvider.GetOAuthClient(tok.AccessToken)
	return listProjects(ctx, client)
}

// FetchRepoPerms is a stub implementation for OAuth authz provider because the API
// endpoint we use requires admin in order to get complete results.
func (p *OAuthAuthzProvider) FetchRepoPerms(context.Context, *api.ExternalRepoSpec) ([]string, error) {
	return []string{}, nil
}

package threads

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater"
)

var (
	cliFactory = repos.NewHTTPClientFactory()
)

func getClientForRepo(ctx context.Context, repoID api.RepoID) (client *github.Client, externalServiceID int64, err error) {
	// 🚨 SECURITY: Only site admins may read external services (they have secrets). TODO!(sqs)
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, 0, err
	}

	svcs, err := repoupdater.DefaultClient.RepoExternalServices(ctx, uint32(repoID))
	if err != nil {
		return nil, 0, err
	}
	// TODO!(sqs): how to choose if there are multiple
	if len(svcs) == 0 {
		return nil, 0, fmt.Errorf("no external services exist for repo %d", repoID)
	}
	src, err := repos.NewGithubSource(&repos.ExternalService{
		ID:          svcs[0].ID,
		Kind:        svcs[0].Kind,
		DisplayName: svcs[0].DisplayName,
		Config:      svcs[0].Config,
		CreatedAt:   svcs[0].CreatedAt,
		UpdatedAt:   svcs[0].UpdatedAt,
	}, cliFactory)
	if err != nil {
		return nil, 0, err
	}
	return src.Client(), svcs[0].ID, nil
}

type githubActor struct {
	Login string `json:"login"`
	URL   string `json:"url"`
}

const githubActorFieldsFragment = `
fragment ActorFields on Actor {
	... on User {
		login
		url
	}
	... on Bot {
		login
		url
	}
}
`

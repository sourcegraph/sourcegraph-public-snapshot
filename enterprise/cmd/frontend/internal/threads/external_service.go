package threads

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/events"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater"
)

type CreateChangesetData struct {
	BaseRefName, HeadRefName string
	Title                    string
	Body                     string

	ExistingThreadID int64
}

type externalServiceClient interface {
	CreateOrUpdateThread(ctx context.Context, repoName api.RepoName, repoID api.RepoID, extRepo api.ExternalRepoSpec, data CreateChangesetData) (threadID int64, err error)
	RefreshThreadMetadata(ctx context.Context, threadID, threadExternalServiceID int64, externalID string, repoID api.RepoID) error
	GetThreadTimelineItems(ctx context.Context, threadExternalID string) ([]events.CreationData, error)
}

var cliFactory = repos.NewHTTPClientFactory()

func getClientForRepo(ctx context.Context, repoID api.RepoID) (client externalServiceClient, externalServiceID int64, err error) {
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
	svc := svcs[0]
	src, err := repos.NewSource(&repos.ExternalService{
		ID:          svc.ID,
		Kind:        svc.Kind,
		DisplayName: svc.DisplayName,
		Config:      svc.Config,
		CreatedAt:   svc.CreatedAt,
		UpdatedAt:   svc.UpdatedAt,
	}, cliFactory)
	if err != nil {
		return nil, 0, err
	}
	switch src := src.(type) {
	case *repos.GithubSource:
		return &githubExternalServiceClient{src: src}, svc.ID, nil
	case *repos.BitbucketServerSource:
		return &bitbucketServerExternalServiceClient{src: src}, svc.ID, nil
	default:
		return nil, 0, fmt.Errorf("unhandled service type %T", src)
	}
}

package webhookhandlers

import (
	"context"
	"fmt"
	gh "github.com/google/go-github/v28/github"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
)

// handleGithubRepoAuthzEvent handles any github event containing a repository field, and enqueues the contained
// repo for permissions synchronisation
func handleGithubRepoAuthzEvent(ctx context.Context, extSvc *repos.ExternalService, payload interface{}) error {
	// TODO: check feature flag

	log15.Warn(fmt.Sprintf("handleGithubRepoAuthzEvent: Got github event %T", payload))

	e, ok := payload.(repoGetter)
	if !ok {
		return fmt.Errorf("incorrect event type sent to github event handler: %T", payload)
	}
	return scheduleRepoUpdate(ctx, e.GetRepo())
}

type repoGetter interface {
	GetRepo() *gh.Repository
}

func scheduleRepoUpdate(ctx context.Context, repo *gh.Repository) error {
	r, err := backend.Repos.GetByName(ctx, api.RepoName("github.com/"+repo.GetFullName()))
	if err != nil {
		return err
	}

	log15.Warn("scheduleRepoUpdate: Dispatching permissions update for repo %s", repo.GetFullName())

	c := repoupdater.DefaultClient
	return c.SchedulePermsSync(ctx, protocol.PermsSyncRequest{
		RepoIDs: []api.RepoID{r.ID},
	})
}

package syncwebhooks

import (
	"context"
	"fmt"

	"github.com/google/go-github/github"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var (
	githubEvents = []string{
		"push",
	}
	Url = ""
)

type SyncGitHubWebhook struct {
}

func (h *SyncGitHubWebhook) Register(router *webhooks.GitHubWebhook) {
	router.Register(
		h.handleSyncWebhook,
		githubEvents...,
	)
}

func (h *SyncGitHubWebhook) handleSyncWebhook(ctx context.Context, extSvc *types.ExternalService, payload any) error {
	fmt.Println("handleSyncWebhook...")
	repo := payload.(*github.PushEvent).GetRepo()
	fmt.Printf("repo:%+v\n", repo)
	// repoName := repo.Name
	var repoName api.RepoName
	repoName = "github.com/sourcegraph/sourcegraph"

	var cli *repoupdater.Client
	if Url == "" {
		cli = repoupdater.DefaultClient
	} else {
		cli = repoupdater.NewClient(Url)
	}

	res, err := cli.EnqueueRepoUpdate(ctx, repoName)
	if err != nil {
		fmt.Println("error in handleSyncWebhook", err)
	}
	fmt.Printf("Enqueued:%+v\n", res)

	return nil
}

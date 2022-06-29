package syncwebhooks

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var (
	githubEvents = []string{
		"push",
	}
	Url string
)

type SyncWebhook struct {
}

func (h *SyncWebhook) Register(router *webhooks.GitHubWebhook) {
	router.Register(
		h.handleSyncWebhook,
		githubEvents...,
	)
}

func (h *SyncWebhook) handleSyncWebhook(ctx context.Context, extSvc *types.ExternalService, payload any) error {
	fmt.Println("handleSyncWebhook...")
	// repo := payload.(*github.PushEvent).GetRepo()
	// fmt.Printf("repo:%+v\n", repo)
	var repo api.RepoName
	repo = "github.com/sourcegraph/sourcegraph"

	cli := repoupdater.NewClient(Url)

	res, err := cli.EnqueueRepoUpdate(ctx, repo)
	if err != nil {
		fmt.Println("error in handleSyncWebhook", err)
	}
	fmt.Printf("res:%+v\n", res)

	return nil
}

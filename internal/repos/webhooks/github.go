package syncwebhooks

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/go-github/v43/github"

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
	name := api.RepoName(*repo.Name)

	var cli *repoupdater.Client
	if Url == "" {
		cli = repoupdater.DefaultClient
	} else {
		cli = repoupdater.NewClient(Url)
	}

	res, err := cli.EnqueueRepoUpdate(ctx, name)
	if err != nil {
		return errors.New(fmt.Sprint("error enqueuing repo", err))
	}
	fmt.Printf("Enqueued:%+v\n", res)

	return nil
}

// u, err := extsvc.WebhookURL(extsvc.TypeGitHub, extSvc.ID, nil, "https://example.com/")
// if err != nil {
// 	t.Fatal(err)
// }

// Indra's answer will affect how I implement the handler in httpapi.go

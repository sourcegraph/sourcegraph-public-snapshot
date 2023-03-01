package webhooks

import (
	"context"
	"fmt"

	"github.com/sourcegraph/log"
	fewebhooks "github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/azuredevops"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var azureDevOpsEvents = []string{
	"git.pullrequest.created",
	"git.pullrequest.updated",
	"git.pullrequest.merged",
}

type AzureDevOpsWebhook struct {
	*webhook
}

func NewAzureDevOpsWebhook(store *store.Store, gitserverClient gitserver.Client, logger log.Logger) *AzureDevOpsWebhook {
	return &AzureDevOpsWebhook{
		webhook: &webhook{store, gitserverClient, logger, extsvc.TypeAzureDevOps},
	}
}

func (h *AzureDevOpsWebhook) Register(router *fewebhooks.Router) {
	router.Register(
		h.handleEvent,
		extsvc.KindAzureDevOps,
		azureDevOpsEvents...,
	)
}

func (h *AzureDevOpsWebhook) handleEvent(ctx context.Context, db database.DB, codeHostURN extsvc.CodeHostBaseURL, event any) error {
	ctx = actor.WithInternalActor(ctx)

	prs, ev, err := h.convertEvent(event)
	if err != nil {
		return err
	}

	for _, pr := range prs {
		if pr == (PR{}) {
			h.logger.Warn("Dropping Azure DevOps webhook event", log.String("type", fmt.Sprintf("%T", event)))
			continue
		}

		eventErr := h.upsertChangesetEvent(ctx, codeHostURN, pr, ev)
		if eventErr != nil {
			err = errors.Append(err, eventErr)
		}
	}
	return err
}

func (h *AzureDevOpsWebhook) convertEvent(theirs interface{}) ([]PR, keyer, error) {
	switch e := theirs.(type) {
	case *azuredevops.PullRequestCreatedEvent:
		return azureDevOpsPullRequestEventPRs(azuredevops.PullRequestEvent(*e)), e, nil
	case *azuredevops.PullRequestMergedEvent:
		return azureDevOpsPullRequestEventPRs(azuredevops.PullRequestEvent(*e)), e, nil
	case *azuredevops.PullRequestUpdatedEvent:
		return azureDevOpsPullRequestEventPRs(azuredevops.PullRequestEvent(*e)), e, nil
	default:
		return nil, nil, errors.Newf("unknown event type: %T", theirs)
	}
}

func azureDevOpsPullRequestEventPRs(e azuredevops.PullRequestEvent) []PR {
	return []PR{
		{
			ID:             int64(e.PullRequest.ID),
			RepoExternalID: e.PullRequest.Repository.ID,
		},
	}
}

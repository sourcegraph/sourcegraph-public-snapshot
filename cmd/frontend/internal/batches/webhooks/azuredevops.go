package webhooks

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sourcegraph/log"

	fewebhooks "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/azuredevops"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var azureDevOpsEvents = []string{
	"git.pullrequest.updated",
	"git.pullrequest.merged",
	"git.pullrequest.approved",
	"git.pullrequest.approved_with_suggestions",
	"git.pullrequest.rejected",
	"git.pullrequest.waiting_for_author",
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

	// If the event is: PullRequestUpdatedEvent or PullRequestMergedEvent, it is best to just not try to derive the state from it and
	// pull it in manually 💪.
	switch e := event.(type) {
	case *azuredevops.PullRequestUpdatedEvent:
		event := azuredevops.PullRequestEvent(*e)
		if err := h.enqueueAzureDevOpsChangesetSyncFromEvent(ctx, codeHostURN, event); err != nil {
			return &httpError{
				code: http.StatusInternalServerError,
				err:  err,
			}
		}
		return nil
	case *azuredevops.PullRequestMergedEvent:
		event := azuredevops.PullRequestEvent(*e)
		if err := h.enqueueAzureDevOpsChangesetSyncFromEvent(ctx, codeHostURN, event); err != nil {
			return &httpError{
				code: http.StatusInternalServerError,
				err:  err,
			}
		}
		return nil
	}
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
	case *azuredevops.PullRequestMergedEvent:
		return azureDevOpsPullRequestEventPRs(azuredevops.PullRequestEvent(*e)), e, nil
	case *azuredevops.PullRequestApprovedEvent:
		return azureDevOpsPullRequestEventPRs(azuredevops.PullRequestEvent(*e)), e, nil
	case *azuredevops.PullRequestApprovedWithSuggestionsEvent:
		return azureDevOpsPullRequestEventPRs(azuredevops.PullRequestEvent(*e)), e, nil
	case *azuredevops.PullRequestWaitingForAuthorEvent:
		return azureDevOpsPullRequestEventPRs(azuredevops.PullRequestEvent(*e)), e, nil
	case *azuredevops.PullRequestRejectedEvent:
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

// enqueueAzureDevOpsChangesetSyncFromEvent enqueues a sync request for a specified changeset in repo-updater.
// This is used instead of deriving the changeset from the incoming webhook event when doing so
// would be difficult.
func (h *AzureDevOpsWebhook) enqueueAzureDevOpsChangesetSyncFromEvent(ctx context.Context, esID extsvc.CodeHostBaseURL, event azuredevops.PullRequestEvent) error {
	// We need to get our changeset ID for this to work. To get _there_, we need
	// the repo ID, and then we can use the merge request ID to match the
	// external ID.
	pr := azureDevOpsToPR(event.PullRequest)
	repo, err := h.getRepoForPR(ctx, h.Store, pr, esID)
	if err != nil {
		return errors.Wrap(err, "getting repo")
	}

	c, err := h.Store.GetChangeset(ctx, store.GetChangesetOpts{
		RepoID:              repo.ID,
		ExternalID:          strconv.FormatInt(pr.ID, 10),
		ExternalServiceType: h.ServiceType,
	})
	if err != nil {
		return errors.Wrap(err, "getting changeset")
	}

	if err := repoupdater.DefaultClient.EnqueueChangesetSync(ctx, []int64{c.ID}); err != nil {
		return errors.Wrap(err, "enqueuing changeset sync")
	}

	return nil
}

// azureDevOpsToPR instantiates a new PR instance given fields that are commonly
// available in Azure DevOps webhook payloads.
func azureDevOpsToPR(pr azuredevops.PullRequest) PR {
	return PR{
		ID:             int64(pr.ID),
		RepoExternalID: pr.Repository.ID,
	}
}

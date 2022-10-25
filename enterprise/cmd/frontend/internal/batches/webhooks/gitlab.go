package webhooks

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/inconshreveable/log15"

	fewebhooks "github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

var (
    gitlabEvents = []string{
        "merge_request",
        "pipeline",
    }
    )

type GitLabWebhook struct {
	*Webhook

	// failHandleEvent is here so that we can explicity force a failure in the event
	// handler in tests
	failHandleEvent error
}

func NewGitLabWebhook(store *store.Store, gitserverClient gitserver.Client) *GitLabWebhook {
	return &GitLabWebhook{Webhook: &Webhook{store, gitserverClient, extsvc.TypeGitLab}}
}

func (h *GitLabWebhook) Register(router *fewebhooks.Webhook) {
    router.Register(
        h.handleEvent,
        extsvc.KindGitLab,
        gitlabEvents...,
    )
}

var (
	errExternalServiceNotFound     = errors.New("external service not found")
	errExternalServiceWrongKind    = errors.New("external service is not of the expected kind")
	errPipelineMissingMergeRequest = errors.New("pipeline event does not include a merge request")
)


// handleEvent is essentially a router: it dispatches based on the event type
// to perform whatever changeset action is appropriate for that event.
func (h *GitLabWebhook) handleEvent(ctx context.Context, db database.DB, codeHostURN extsvc.CodeHostBaseURL, event any) error {
	log15.Debug("GitLab webhook received", "type", fmt.Sprintf("%T", event))

	if h.failHandleEvent != nil {
        return h.failHandleEvent
	}

	switch e := event.(type) {
	// Some merge request event types require us to do a full resync.
	//
	// For example, approvals and unapprovals manifest in normal syncs as
	// system notes, but we _don't_ get them as note events in webhooks.
	// Instead, we get a merge request webhook with an "approved" or
	// "unapproved" action field that magically appears in the object (merge
	// request) attributes and no further details on who changed the approval
	// status, the note ID, or anything else we can use to deduplicate later.
	//
	// Similarly, for update events, we don't get the full set of fields that
	// we get when we sync using the REST API (presumably because this reflects
	// the data types at the point webhooks were added to GitLab several years
	// ago, and not today): labels come in a different format outside of the
	// merge request, and we'd still have to go query for notes and pipelines.
	//
	// Therefore, the only realistic action we can take here is to re-sync the
	// changeset as a whole. The problem is that — since we only have the merge
	// request — this requires three requests to the REST API, and GitLab's
	// documentation is quite clear that webhooks should run as fast as possible
	// to avoid unexpected retries.
	//
	// To meet this goal, rather than synchronously synchronizing here, we'll
	// instead ask repo-updater to prioritize the sync of this changeset and let
	// the normal sync process take care of pulling the notes and pipelines and
	// putting things in the right places. The downside is that the updated
	// changeset state won't appear _quite_ as instantaneously to the user, but
	// this is the best compromise given the limited payload we get in the
	// webhook.
	case *webhooks.MergeRequestApprovedEvent,
		*webhooks.MergeRequestUnapprovedEvent,
		*webhooks.MergeRequestUpdateEvent:
		if err := h.enqueueChangesetSyncFromEvent(ctx, codeHostURN, e.(webhooks.MergeRequestEventCommonContainer).ToEventCommon()); err != nil {
			return &httpError{
				code: http.StatusInternalServerError,
				err:  err,
			}
		}
		return nil

	case webhooks.UpsertableWebhookEvent:
		eventCommon := e.ToEventCommon()
		event := e.ToEvent()
		pr := gitlabToPR(&eventCommon.Project, eventCommon.MergeRequest)
		if err := h.upsertChangesetEvent(ctx, codeHostURN, pr, event); err != nil {
			return &httpError{
				code: http.StatusInternalServerError,
				err:  errors.Wrap(err, "upserting changeset event"),
			}
		}
		return nil

	case *webhooks.PipelineEvent:
		if err := h.handlePipelineEvent(ctx, codeHostURN, e); err != nil && err != errPipelineMissingMergeRequest {
			return &httpError{
				code: http.StatusInternalServerError,
				err:  err,
			}
		}
		return nil
	}

	// We don't want to return a non-2XX status code and have GitLab retry the
	// webhook, so we'll log that we don't know what to do and return 204.
	log15.Debug("cannot handle GitLab webhook event of unknown type", "event", event, "type", fmt.Sprintf("%T", event))
	return nil
}

func (h *GitLabWebhook) enqueueChangesetSyncFromEvent(ctx context.Context, esID extsvc.CodeHostBaseURL, event *webhooks.MergeRequestEventCommon) error {
	// We need to get our changeset ID for this to work. To get _there_, we need
	// the repo ID, and then we can use the merge request IID to match the
	// external ID.
	pr := gitlabToPR(&event.Project, event.MergeRequest)
	repo, err := h.getRepoForPR(ctx, h.Store, pr, esID)
	if err != nil {
		return errors.Wrap(err, "getting repo")
	}

	c, err := h.getChangesetForPR(ctx, h.Store, &pr, repo)
	if err != nil {
		return errors.Wrap(err, "getting changeset")
	}

	if err := repoupdater.DefaultClient.EnqueueChangesetSync(ctx, []int64{c.ID}); err != nil {
		return errors.Wrap(err, "enqueuing changeset sync")
	}

	return nil
}

func (h *GitLabWebhook) handlePipelineEvent(ctx context.Context, esID extsvc.CodeHostBaseURL, event *webhooks.PipelineEvent) error {
	// Pipeline webhook payloads don't include the merge request very reliably:
	// for example, re-running a pipeline from the GitLab UI will result in no
	// merge request field, even when that pipeline was attached to a merge
	// request. So the very first thing we need to do is see if we even have the
	// merge request; if we don't, we can't do anything useful here, and we'll
	// just have to wait for the next scheduled sync.
	if event.MergeRequest == nil {
		log15.Debug("ignoring pipeline event without a merge request", "payload", event)
		return errPipelineMissingMergeRequest
	}

	pr := gitlabToPR(&event.Project, event.MergeRequest)
	if err := h.upsertChangesetEvent(ctx, esID, pr, &event.Pipeline); err != nil {
		return errors.Wrap(err, "upserting changeset event")
	}
	return nil
}

func (h *GitLabWebhook) getChangesetForPR(ctx context.Context, tx *store.Store, pr *PR, repo *types.Repo) (*btypes.Changeset, error) {
	return tx.GetChangeset(ctx, store.GetChangesetOpts{
		RepoID:              repo.ID,
		ExternalID:          strconv.FormatInt(pr.ID, 10),
		ExternalServiceType: h.ServiceType,
	})
}

// gitlabToPR instantiates a new PR instance given fields that are commonly
// available in GitLab webhook payloads.
func gitlabToPR(project *gitlab.ProjectCommon, mr *gitlab.MergeRequest) PR {
	return PR{
		ID:             int64(mr.IID),
		RepoExternalID: strconv.Itoa(project.ID),
	}
}

// validateGitLabSecret validates that the given secret matches one of the
// webhooks in the external service.
func validateGitLabSecret(ctx context.Context, extSvc *types.ExternalService, secret string) (bool, error) {
	// An empty secret never succeeds.
	if secret == "" {
		return false, nil
	}

	// Get the typed configuration.
	c, err := extSvc.Configuration(ctx)
	if err != nil {
		return false, errors.Wrap(err, "getting external service configuration")
	}

	config, ok := c.(*schema.GitLabConnection)
	if !ok {
		return false, errExternalServiceWrongKind
	}

	// Iterate over the webhooks and look for one with the right secret. The
	// number of webhooks in an external service should be small enough that a
	// linear search like this is sufficient.
	for _, webhook := range config.Webhooks {
		if webhook.Secret == secret {
			return true, nil
		}
	}
	return false, nil
}

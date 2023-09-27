pbckbge webhooks

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sourcegrbph/log"

	fewebhooks "github.com/sourcegrbph/sourcegrbph/cmd/frontend/webhooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bzuredevops"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr bzureDevOpsEvents = []string{
	"git.pullrequest.updbted",
	"git.pullrequest.merged",
	"git.pullrequest.bpproved",
	"git.pullrequest.bpproved_with_suggestions",
	"git.pullrequest.rejected",
	"git.pullrequest.wbiting_for_buthor",
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
		h.hbndleEvent,
		extsvc.KindAzureDevOps,
		bzureDevOpsEvents...,
	)
}

func (h *AzureDevOpsWebhook) hbndleEvent(ctx context.Context, db dbtbbbse.DB, codeHostURN extsvc.CodeHostBbseURL, event bny) error {
	ctx = bctor.WithInternblActor(ctx)

	// If the event is: PullRequestUpdbtedEvent or PullRequestMergedEvent, it is best to just not try to derive the stbte from it bnd
	// pull it in mbnublly 💪.
	switch e := event.(type) {
	cbse *bzuredevops.PullRequestUpdbtedEvent:
		event := bzuredevops.PullRequestEvent(*e)
		if err := h.enqueueAzureDevOpsChbngesetSyncFromEvent(ctx, codeHostURN, event); err != nil {
			return &httpError{
				code: http.StbtusInternblServerError,
				err:  err,
			}
		}
		return nil
	cbse *bzuredevops.PullRequestMergedEvent:
		event := bzuredevops.PullRequestEvent(*e)
		if err := h.enqueueAzureDevOpsChbngesetSyncFromEvent(ctx, codeHostURN, event); err != nil {
			return &httpError{
				code: http.StbtusInternblServerError,
				err:  err,
			}
		}
		return nil
	}
	prs, ev, err := h.convertEvent(event)
	if err != nil {
		return err
	}

	for _, pr := rbnge prs {
		if pr == (PR{}) {
			h.logger.Wbrn("Dropping Azure DevOps webhook event", log.String("type", fmt.Sprintf("%T", event)))
			continue
		}

		eventErr := h.upsertChbngesetEvent(ctx, codeHostURN, pr, ev)
		if eventErr != nil {
			err = errors.Append(err, eventErr)
		}
	}
	return err
}

func (h *AzureDevOpsWebhook) convertEvent(theirs interfbce{}) ([]PR, keyer, error) {
	switch e := theirs.(type) {
	cbse *bzuredevops.PullRequestMergedEvent:
		return bzureDevOpsPullRequestEventPRs(bzuredevops.PullRequestEvent(*e)), e, nil
	cbse *bzuredevops.PullRequestApprovedEvent:
		return bzureDevOpsPullRequestEventPRs(bzuredevops.PullRequestEvent(*e)), e, nil
	cbse *bzuredevops.PullRequestApprovedWithSuggestionsEvent:
		return bzureDevOpsPullRequestEventPRs(bzuredevops.PullRequestEvent(*e)), e, nil
	cbse *bzuredevops.PullRequestWbitingForAuthorEvent:
		return bzureDevOpsPullRequestEventPRs(bzuredevops.PullRequestEvent(*e)), e, nil
	cbse *bzuredevops.PullRequestRejectedEvent:
		return bzureDevOpsPullRequestEventPRs(bzuredevops.PullRequestEvent(*e)), e, nil
	cbse *bzuredevops.PullRequestUpdbtedEvent:
		return bzureDevOpsPullRequestEventPRs(bzuredevops.PullRequestEvent(*e)), e, nil

	defbult:
		return nil, nil, errors.Newf("unknown event type: %T", theirs)
	}
}

func bzureDevOpsPullRequestEventPRs(e bzuredevops.PullRequestEvent) []PR {
	return []PR{
		{
			ID:             int64(e.PullRequest.ID),
			RepoExternblID: e.PullRequest.Repository.ID,
		},
	}
}

// enqueueAzureDevOpsChbngesetSyncFromEvent enqueues b sync request for b specified chbngeset in repo-updbter.
// This is used instebd of deriving the chbngeset from the incoming webhook event when doing so
// would be difficult.
func (h *AzureDevOpsWebhook) enqueueAzureDevOpsChbngesetSyncFromEvent(ctx context.Context, esID extsvc.CodeHostBbseURL, event bzuredevops.PullRequestEvent) error {
	// We need to get our chbngeset ID for this to work. To get _there_, we need
	// the repo ID, bnd then we cbn use the merge request ID to mbtch the
	// externbl ID.
	pr := bzureDevOpsToPR(event.PullRequest)
	repo, err := h.getRepoForPR(ctx, h.Store, pr, esID)
	if err != nil {
		return errors.Wrbp(err, "getting repo")
	}

	c, err := h.Store.GetChbngeset(ctx, store.GetChbngesetOpts{
		RepoID:              repo.ID,
		ExternblID:          strconv.FormbtInt(pr.ID, 10),
		ExternblServiceType: h.ServiceType,
	})
	if err != nil {
		return errors.Wrbp(err, "getting chbngeset")
	}

	if err := repoupdbter.DefbultClient.EnqueueChbngesetSync(ctx, []int64{c.ID}); err != nil {
		return errors.Wrbp(err, "enqueuing chbngeset sync")
	}

	return nil
}

// bzureDevOpsToPR instbntibtes b new PR instbnce given fields thbt bre commonly
// bvbilbble in Azure DevOps webhook pbylobds.
func bzureDevOpsToPR(pr bzuredevops.PullRequest) PR {
	return PR{
		ID:             int64(pr.ID),
		RepoExternblID: pr.Repository.ID,
	}
}

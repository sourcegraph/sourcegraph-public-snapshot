package webhooks

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/inconshreveable/log15"

	fewebhooks "github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type BitbucketCloudWebhook struct {
	*Webhook
}

func NewBitbucketCloudWebhook(store *store.Store) *BitbucketCloudWebhook {
	return &BitbucketCloudWebhook{
		Webhook: &Webhook{store, extsvc.TypeBitbucketCloud},
	}
}

func (h *BitbucketCloudWebhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e, extSvc, hErr := h.parseEvent(r)
	if hErr != nil {
		respond(w, hErr.code, hErr)
		return
	}

	fewebhooks.SetExternalServiceID(r.Context(), extSvc.ID)

	// 🚨 SECURITY: now that the shared secret has been validated, we can use an
	// internal actor on the context.
	ctx := actor.WithInternalActor(r.Context())

	externalServiceID, err := extractExternalServiceID(extSvc)
	if err != nil {
		respond(w, http.StatusInternalServerError, err)
		return
	}

	prs, ev, err := h.convertEvent(e)
	if err != nil {
		if !errors.Is(err, bitbucketcloud.UnknownWebhookEventKey("")) {
			respond(w, http.StatusInternalServerError, err)
		} else {
			// Unknown type, so we'll just ignore it.
			return
		}
	}

	var m error
	for _, pr := range prs {
		log15.Info("processing", "pr", pr, "ev", fmt.Sprintf("%+v", ev))
		err := h.upsertChangesetEvent(ctx, externalServiceID, pr, ev)
		if err != nil {
			m = errors.Append(m, err)
		}
	}
	if m != nil {
		respond(w, http.StatusInternalServerError, m)
	}
}

func (h *BitbucketCloudWebhook) parseEvent(r *http.Request) (interface{}, *types.ExternalService, *httpError) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, &httpError{http.StatusInternalServerError, err}
	}

	rawID := r.FormValue(extsvc.IDParam)
	var externalServiceID int64
	// id could be blank temporarily if we haven't updated the hook url to include the param yet
	if rawID != "" {
		externalServiceID, err = strconv.ParseInt(rawID, 10, 64)
		if err != nil {
			return nil, nil, &httpError{http.StatusBadRequest, errors.Wrap(err, "invalid external service id")}
		}
	}

	args := database.ExternalServicesListOptions{Kinds: []string{extsvc.KindBitbucketCloud}}
	if externalServiceID != 0 {
		args.IDs = append(args.IDs, externalServiceID)
	}
	es, err := h.Store.ExternalServices().List(r.Context(), args)
	if err != nil {
		return nil, nil, &httpError{http.StatusInternalServerError, err}
	}

	var extSvc *types.ExternalService
	for _, e := range es {
		if externalServiceID != 0 && e.ID != externalServiceID {
			continue
		}

		c, _ := e.Configuration()
		con, ok := c.(*schema.BitbucketCloudConnection)
		if !ok {
			continue
		}

		if secret := con.WebhookSecret; secret != "" {
			if r.FormValue("secret") == secret {
				extSvc = e
				break
			}
		}
	}

	if extSvc == nil || err != nil {
		return nil, nil, &httpError{http.StatusUnauthorized, err}
	}

	e, err := bitbucketcloud.ParseWebhookEvent(r.Header.Get("X-Event-Key"), payload)
	if err != nil {
		return nil, nil, &httpError{http.StatusBadRequest, errors.Wrap(err, "parsing webhook")}
	}

	return e, extSvc, nil
}

func (h *BitbucketCloudWebhook) convertEvent(theirs interface{}) ([]PR, keyer, error) {
	switch e := theirs.(type) {
	case *bitbucketcloud.PullRequestApprovedEvent:
		return bitbucketCloudPullRequestEventPRs(&e.PullRequestEvent), e, nil
	case *bitbucketcloud.PullRequestChangesRequestCreatedEvent:
		return bitbucketCloudPullRequestEventPRs(&e.PullRequestEvent), e, nil
	case *bitbucketcloud.PullRequestChangesRequestRemovedEvent:
		return bitbucketCloudPullRequestEventPRs(&e.PullRequestEvent), e, nil
	case *bitbucketcloud.PullRequestCommentCreatedEvent:
		return bitbucketCloudPullRequestEventPRs(&e.PullRequestEvent), e, nil
	case *bitbucketcloud.PullRequestCommentDeletedEvent:
		return bitbucketCloudPullRequestEventPRs(&e.PullRequestEvent), e, nil
	case *bitbucketcloud.PullRequestCommentUpdatedEvent:
		return bitbucketCloudPullRequestEventPRs(&e.PullRequestEvent), e, nil
	case *bitbucketcloud.PullRequestFulfilledEvent:
		return bitbucketCloudPullRequestEventPRs(&e.PullRequestEvent), e, nil
	case *bitbucketcloud.PullRequestRejectedEvent:
		return bitbucketCloudPullRequestEventPRs(&e.PullRequestEvent), e, nil
	case *bitbucketcloud.PullRequestUnapprovedEvent:
		return bitbucketCloudPullRequestEventPRs(&e.PullRequestEvent), e, nil
	case *bitbucketcloud.PullRequestUpdatedEvent:
		return bitbucketCloudPullRequestEventPRs(&e.PullRequestEvent), e, nil
	case *bitbucketcloud.RepoCommitStatusCreatedEvent,
		*bitbucketcloud.RepoCommitStatusUpdatedEvent:
		// TODO: figure out how the fuck to get the head commit out of a
		// changeset and see if we have a match, since we _should_ get a
		// pullrequest:updated before any commit statuses.
		return nil, nil, errors.New("unimplemented")
	default:
		return nil, nil, errors.Newf("unknown event type: %T", theirs)
	}
}

func bitbucketCloudPullRequestEventPRs(e *bitbucketcloud.PullRequestEvent) []PR {
	return []PR{
		{
			ID:             int64(e.PullRequest.ID),
			RepoExternalID: e.Repository.UUID,
		},
	}
}

package webhooks

import (
	"context"
	"io"
	"net/http"
	"strconv"

	fewebhooks "github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	bbcs "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
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

	prs, ev, err := h.convertEvent(r.Context(), e, externalServiceID)
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
		err := h.upsertChangesetEvent(ctx, externalServiceID, pr, ev)
		if err != nil {
			m = errors.Append(m, err)
		}
	}
	if m != nil {
		respond(w, http.StatusInternalServerError, m)
	} else {
		respond(w, http.StatusNoContent, nil)
	}
}

func (h *BitbucketCloudWebhook) parseEvent(r *http.Request) (interface{}, *types.ExternalService, *httpError) {
	if r.Body == nil {
		return nil, nil, &httpError{http.StatusBadRequest, errors.New("nil request body")}
	}

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

func (h *BitbucketCloudWebhook) convertEvent(ctx context.Context, theirs interface{}, externalServiceID string) ([]PR, keyer, error) {
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
	case *bitbucketcloud.RepoCommitStatusCreatedEvent:
		prs, err := bitbucketCloudRepoCommitStatusEventPRs(ctx, h.Store, &e.RepoCommitStatusEvent, externalServiceID)
		return prs, e, err
	case *bitbucketcloud.RepoCommitStatusUpdatedEvent:
		prs, err := bitbucketCloudRepoCommitStatusEventPRs(ctx, h.Store, &e.RepoCommitStatusEvent, externalServiceID)
		return prs, e, err
	default:
		return nil, nil, errors.Newf("unknown event type: %T", theirs)
	}
}

func bitbucketCloudPullRequestEventPRs(e *bitbucketcloud.PullRequestEvent) []PR {
	return []PR{
		{
			ID:             e.PullRequest.ID,
			RepoExternalID: e.Repository.UUID,
		},
	}
}

func bitbucketCloudRepoCommitStatusEventPRs(
	ctx context.Context, bstore *store.Store,
	e *bitbucketcloud.RepoCommitStatusEvent, externalServiceID string,
) ([]PR, error) {
	// Bitbucket Cloud repo commit statuses only include the commit hash they
	// relate to, not the branch or PR, so we have to go look up the relevant
	// changeset(s) from the database.

	// First up, let's find the repos ID so we can limit the changeset search.
	repos, err := bstore.Repos().List(ctx, database.ReposListOptions{
		ExternalRepos: []api.ExternalRepoSpec{
			{
				ID:          e.Repository.UUID,
				ServiceType: extsvc.TypeBitbucketCloud,
				ServiceID:   externalServiceID,
			},
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "cannot find repo with ID=%q ServiceType=%q ServiceID=%q", e.Repository.UUID, extsvc.TypeBitbucketCloud, externalServiceID)
	}
	if len(repos) != 1 {
		return nil, errors.Wrapf(err, "unexpected number of repos matched: %d", len(repos))
	}
	repo := repos[0]

	// Now we can look up the changeset(s).
	changesets, _, err := bstore.ListChangesets(ctx, store.ListChangesetsOpts{
		BitbucketCloudCommit: e.CommitStatus.Commit.Hash,
		RepoID:               repo.ID,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "listing changesets matched to repo ID=%d", repo.ID)
	}

	prs := make([]PR, len(changesets))
	for i, changeset := range changesets {
		prs[i] = PR{
			ID:             changeset.Metadata.(*bbcs.AnnotatedPullRequest).ID,
			RepoExternalID: e.Repository.UUID,
		}
	}
	return prs, nil
}

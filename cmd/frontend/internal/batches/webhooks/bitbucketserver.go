package webhooks

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	gh "github.com/google/go-github/v55/github"
	sglog "github.com/sourcegraph/log"

	fewebhooks "github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

var bitbucketServerEvents = []string{
	"ping",
	"repo:build_status",
	"pr:activity:status",
	"pr:activity:event",
	"pr:activity:rescope",
	"pr:activity:merge",
	"pr:activity:comment",
	"pr:activity:reviewers",
	"pr:participant:status",
}

type BitbucketServerWebhook struct {
	*webhook
}

func NewBitbucketServerWebhook(store *store.Store, gitserverClient gitserver.Client, logger sglog.Logger) *BitbucketServerWebhook {
	return &BitbucketServerWebhook{
		webhook: &webhook{store, gitserverClient, logger, extsvc.TypeBitbucketServer},
	}
}

func (h *BitbucketServerWebhook) Register(router *fewebhooks.Router) {
	router.Register(
		h.handleEvent,
		extsvc.KindBitbucketServer,
		bitbucketServerEvents...,
	)
}

func (h *BitbucketServerWebhook) handleEvent(ctx context.Context, db database.DB, codeHostURN extsvc.CodeHostBaseURL, event any) error {
	// 🚨 SECURITY: If we've made it here, then the secret for the Bitbucket Server webhook has been validated, so we can use
	// an internal actor on the context.
	ctx = actor.WithInternalActor(ctx)

	prs, ev := h.convertEvent(event)

	var err error
	for _, pr := range prs {
		if pr == (PR{}) {
			h.logger.Warn("Dropping Bitbucket Server webhook event", sglog.String("type", fmt.Sprintf("%T", event)))
			continue
		}

		eventError := h.upsertChangesetEvent(ctx, codeHostURN, pr, ev)
		if eventError != nil {
			err = errors.Append(err, eventError)
		}
	}
	return err
}

func (h *BitbucketServerWebhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e, extSvc, hErr := h.parseEvent(r)
	if hErr != nil {
		respond(w, hErr.code, hErr)
		return
	}

	fewebhooks.SetExternalServiceID(r.Context(), extSvc.ID)

	// 🚨 SECURITY: now that the shared secret has been validated, we can use an
	// internal actor on the context.
	ctx := actor.WithInternalActor(r.Context())

	c, err := extSvc.Configuration(r.Context())
	if err != nil {
		h.logger.Error("Could not decode external service config", sglog.Error(err))
		http.Error(w, "Invalid external service config", http.StatusInternalServerError)
		return
	}

	config, ok := c.(*schema.BitbucketServerConnection)
	if !ok {
		h.logger.Error("Could not decode external service config")
		http.Error(w, "Invalid external service config", http.StatusInternalServerError)
		return
	}

	codeHostURN, err := extsvc.NewCodeHostBaseURL(config.Url)
	if err != nil {
		respond(w, http.StatusInternalServerError, errors.Wrap(err, "parsing code host base url"))
	}
	m := h.handleEvent(ctx, h.Store.DatabaseDB(), codeHostURN, e)
	if m != nil {
		respond(w, http.StatusInternalServerError, m)
	}
}

func (h *BitbucketServerWebhook) parseEvent(r *http.Request) (any, *types.ExternalService, *httpError) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, &httpError{http.StatusInternalServerError, err}
	}

	sig := r.Header.Get("X-Hub-Signature")

	rawID := r.FormValue(extsvc.IDParam)
	var externalServiceID int64
	// id could be blank temporarily if we haven't updated the hook url to include the param yet
	if rawID != "" {
		externalServiceID, err = strconv.ParseInt(rawID, 10, 64)
		if err != nil {
			return nil, nil, &httpError{http.StatusBadRequest, errors.Wrap(err, "invalid external service id")}
		}
	}

	args := database.ExternalServicesListOptions{Kinds: []string{extsvc.KindBitbucketServer}}
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

		c, _ := e.Configuration(r.Context())
		con, ok := c.(*schema.BitbucketServerConnection)
		if !ok {
			continue
		}

		if secret := con.WebhookSecret(); secret != "" {
			if err = gh.ValidateSignature(sig, payload, []byte(secret)); err == nil {
				extSvc = e
				break
			}
		}
	}

	if extSvc == nil || err != nil {
		return nil, nil, &httpError{http.StatusUnauthorized, err}
	}

	e, err := bitbucketserver.ParseWebhookEvent(bitbucketserver.WebhookEventType(r), payload)
	if err != nil {
		return nil, nil, &httpError{http.StatusBadRequest, errors.Wrap(err, "parsing webhook")}
	}
	return e, extSvc, nil
}

func (h *BitbucketServerWebhook) convertEvent(theirs any) (prs []PR, ours keyer) {
	h.logger.Debug("Bitbucket Server webhook received", sglog.String("type", fmt.Sprintf("%T", theirs)))

	switch e := theirs.(type) {
	case *bitbucketserver.PullRequestActivityEvent:
		repoID := strconv.Itoa(e.PullRequest.FromRef.Repository.ID)
		pr := PR{ID: int64(e.PullRequest.ID), RepoExternalID: repoID}
		prs = append(prs, pr)
		return prs, e.Activity
	case *bitbucketserver.PullRequestParticipantStatusEvent:
		repoID := strconv.Itoa(e.PullRequest.FromRef.Repository.ID)
		pr := PR{ID: int64(e.PullRequest.ID), RepoExternalID: repoID}
		prs = append(prs, pr)
		return prs, e.ParticipantStatusEvent
	case *bitbucketserver.BuildStatusEvent:
		for _, p := range e.PullRequests {
			repoID := strconv.Itoa(p.FromRef.Repository.ID)
			pr := PR{ID: int64(p.ID), RepoExternalID: repoID}
			prs = append(prs, pr)
		}
		return prs, &bitbucketserver.CommitStatus{
			Commit: e.Commit,
			Status: e.Status,
		}
	}

	return
}

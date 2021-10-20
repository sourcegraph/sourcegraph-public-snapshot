package webhooks

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/cockroachdb/errors"
	gh "github.com/google/go-github/v28/github"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

type BitbucketServerWebhook struct {
	*Webhook
}

func NewBitbucketServerWebhook(store *store.Store) *BitbucketServerWebhook {
	return &BitbucketServerWebhook{
		Webhook: &Webhook{store, extsvc.TypeBitbucketServer},
	}
}

func (h *BitbucketServerWebhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e, extSvc, hErr := h.parseEvent(r)
	if hErr != nil {
		respond(w, hErr.code, hErr)
		return
	}

	// 🚨 SECURITY: now that the shared secret has been validated, we can use an
	// internal actor on the context.
	ctx := actor.WithInternalActor(r.Context())

	externalServiceID, err := extractExternalServiceID(extSvc)
	if err != nil {
		respond(w, http.StatusInternalServerError, err)
		return
	}

	prs, ev := h.convertEvent(e)

	m := new(multierror.Error)
	for _, pr := range prs {
		if pr == (PR{}) {
			log15.Warn("Dropping Bitbucket Server webhook event", "type", fmt.Sprintf("%T", e))
			continue
		}

		err := h.upsertChangesetEvent(ctx, externalServiceID, pr, ev)
		if err != nil {
			m = multierror.Append(m, err)
		}
	}
	if m.ErrorOrNil() != nil {
		respond(w, http.StatusInternalServerError, m)
	}
}

func (h *BitbucketServerWebhook) parseEvent(r *http.Request) (interface{}, *types.ExternalService, *httpError) {
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

		c, _ := e.Configuration()
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

func (h *BitbucketServerWebhook) convertEvent(theirs interface{}) (prs []PR, ours keyer) {
	log15.Debug("Bitbucket Server webhook received", "type", fmt.Sprintf("%T", theirs))

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

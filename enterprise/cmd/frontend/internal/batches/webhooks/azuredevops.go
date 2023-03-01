package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	sglog "github.com/sourcegraph/log"

	fewebhooks "github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/azuredevops"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

var azureDevOpsEvents = []string{
	"git.pullrequest.created",
	"git.pullrequest.updated",
	"git.pullrequest.merged",
}

type AzureDevOpsWebhook struct {
	*webhook
}

func NewAzureDevOpsWebhook(store *store.Store, gitserverClient gitserver.Client, logger sglog.Logger) *AzureDevOpsWebhook {
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
			h.logger.Warn("Dropping Azure DevOps webhook event", sglog.String("type", fmt.Sprintf("%T", event)))
			continue
		}

		eventErr := h.upsertChangesetEvent(ctx, codeHostURN, pr, ev)
		if eventErr != nil {
			err = errors.Append(err, eventErr)
		}
	}
	return err
}

func (h *AzureDevOpsWebhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e, extSvc, hErr := h.parseEvent(r)
	if hErr != nil {
		respond(w, hErr.code, hErr)
		return
	}

	fewebhooks.SetExternalServiceID(r.Context(), extSvc.ID)

	// 🚨 SECURITY: now that the shared secret has been validated, we can use an
	// internal actor on the context.
	ctx := actor.WithInternalActor(r.Context())

	c, err := extSvc.Configuration(ctx)
	if err != nil {
		h.logger.Error("Could not decode external service config", sglog.Error(err))
		http.Error(w, "Invalid external service config", http.StatusInternalServerError)
		return
	}

	config, ok := c.(*schema.AzureDevOpsConnection)
	if !ok {
		h.logger.Error("Could not decode external service config", sglog.Error(err))
		http.Error(w, "Invalid external service config", http.StatusInternalServerError)
		return
	}

	codeHostURN, err := extsvc.NewCodeHostBaseURL(config.Url)
	if err != nil {
		respond(w, http.StatusInternalServerError, errors.Wrap(err, "parsing code host base url"))
	}
	err = h.handleEvent(ctx, h.Store.DatabaseDB(), codeHostURN, e)
	if err != nil {
		respond(w, http.StatusInternalServerError, err)
	} else {
		respond(w, http.StatusNoContent, nil)
	}
}

func (h *AzureDevOpsWebhook) parseEvent(r *http.Request) (interface{}, *types.ExternalService, *httpError) {
	if r.Body == nil {
		return nil, nil, &httpError{http.StatusBadRequest, errors.New("nil request body")}
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, &httpError{http.StatusInternalServerError, err}
	}

	extSvc, err := h.getExternalServiceFromRawID(r.Context(), r.FormValue(extsvc.IDParam))
	if err != nil {
		return nil, nil, &httpError{http.StatusInternalServerError, err}
	}

	if extSvc == nil || err != nil {
		return nil, nil, &httpError{http.StatusUnauthorized, err}
	}

	var event azuredevops.BaseEvent
	err = json.Unmarshal(payload, &event)
	if err != nil {
		return nil, nil, &httpError{http.StatusInternalServerError, err}
	}
	e, err := azuredevops.ParseWebhookEvent(event.EventType, payload)
	if err != nil {
		return nil, nil, &httpError{http.StatusBadRequest, errors.Wrap(err, "parsing webhook")}
	}

	return e, extSvc, nil
}

func (h *AzureDevOpsWebhook) getExternalServiceFromRawID(ctx context.Context, raw string) (*types.ExternalService, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "parsing the raw external service ID")
	}

	es, err := h.Store.ExternalServices().List(ctx, database.ExternalServicesListOptions{
		IDs:   []int64{id},
		Kinds: []string{extsvc.KindGitLab},
	})
	if err != nil {
		return nil, errors.Wrap(err, "listing external services")
	}

	if len(es) == 0 {
		return nil, errExternalServiceNotFound
	} else if len(es) > 1 {
		// This _really_ shouldn't happen, since we provided only one ID above.
		return nil, errors.New("too many external services found")
	}

	return es[0], nil
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

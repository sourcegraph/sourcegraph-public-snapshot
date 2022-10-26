package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/sync/errgroup"
)

type GitLabWebhook Webhook

var (
	errExternalServiceNotFound     = errors.New("external service not found")
	errExternalServiceWrongKind    = errors.New("external service is not of the expected kind")
	errPipelineMissingMergeRequest = errors.New("pipeline event does not include a merge request")
)

func (h *GitLabWebhook) HandleWebhook(w http.ResponseWriter, r *http.Request, codeHostURN extsvc.CodeHostBaseURL) {

	// 🚨 SECURITY: now that the shared secret has been validated, we can use an
	// internal actor on the context.
	ctx := actor.WithInternalActor(r.Context())

	// Parse the event proper.
	if r.Body == nil {
		http.Error(w, "missing request body", http.StatusBadRequest)
		return
	}
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, errors.Wrap(err, "reading payload").Error(), http.StatusInternalServerError)
		return
	}

	var eventKind struct {
		ObjectKind string `json:"object_kind"`
	}
	if err := json.Unmarshal(payload, &eventKind); err != nil {
        http.Error(w, errors.Wrap(err, "determining object kind").Error(), http.StatusInternalServerError)
        return
	}

	event, err := webhooks.UnmarshalEvent(payload)
	if err != nil {
		if errors.Is(err, webhooks.ErrObjectKindUnknown) {
			// We don't want to return a non-2XX status code and have GitLab
			// retry the webhook, so we'll log that we don't know what to do
			// and return 204.
			log15.Debug("unknown object kind", "err", err)

			// We don't use respond() here so that we don't log an error, since
			// this really isn't one.
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusNoContent)
			fmt.Fprintf(w, "%v", err)
		} else {
			http.Error(w, errors.Wrap(err, "unmarshalling payload").Error(), http.StatusInternalServerError)
		}
		return
	}

	// Route the request based on the event type.
    err = h.Dispatch(ctx, eventKind.ObjectKind, codeHostURN, event)
	if err != nil {
		log15.Error("Error handling github webhook event", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *GitLabWebhook) Dispatch(ctx context.Context, eventType string, codeHostURN extsvc.CodeHostBaseURL, e any) error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	g := errgroup.Group{}
	for _, handler := range h.handlers[extsvc.KindGitLab][eventType] {
		// capture the handler variable within this loop
		handler := handler
		g.Go(func() error {
			return handler(ctx, h.DB, codeHostURN, e)
		})
	}
	return g.Wait()
}

// getExternalServiceFromRawID retrieves the external service matching the
// given raw ID, which is usually going to be the string in the
// externalServiceID URL parameter.
//
// On failure, errExternalServiceNotFound is returned if the ID doesn't match
// any GitLab service.
func (h *GitLabWebhook) getExternalServiceFromRawID(ctx context.Context, raw string) (*types.ExternalService, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "parsing the raw external service ID")
	}

	es, err := h.DB.ExternalServices().List(ctx, database.ExternalServicesListOptions{
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

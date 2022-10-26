package webhooks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab/webhooks"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitLabWebhook struct {
	*WebhookRouter
}

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
	err = h.Dispatch(ctx, eventKind.ObjectKind, extsvc.KindGitLab, codeHostURN, event)
	if err != nil {
		log15.Error("Error handling github webhook event", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

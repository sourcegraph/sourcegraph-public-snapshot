package webhooks

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab/webhooks"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (wr *Router) HandleGitLabWebhook(ctx context.Context, logger log.Logger, w http.ResponseWriter, codeHostURN extsvc.CodeHostBaseURL, payload []byte) {
	// 🚨 SECURITY: now that the shared secret has been validated, we can use an
	// internal actor on the context.
	ctx = actor.WithInternalActor(ctx)

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
			logger.Debug("unknown object kind", log.Error(err))

			// We don't use respond() here so that we don't log an error, since
			// this really isn't one.
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusNoContent)
			fmt.Fprintf(w, "%v", err)
		} else {
			http.Error(w, errors.Wrap(err, "unmarshalling event kind from webhook payload").Error(), http.StatusInternalServerError)
		}
		return
	}

	// Route the request based on the event type.
	err = wr.Dispatch(ctx, eventKind.ObjectKind, extsvc.KindGitLab, codeHostURN, event)
	if err != nil {
		logger.Error("Error handling gitlab webhook event", log.Error(err))
		if errcode.IsNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func gitlabValidatePayload(r *http.Request, secret string) ([]byte, error) {
	glSecret := r.Header.Get("X-Gitlab-Token")
	if subtle.ConstantTimeCompare([]byte(glSecret), []byte(secret)) != 1 {
		return nil, errors.New("secrets don't match!")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return body, errors.Wrap(r.Body.Close(), "closing body")
}

func (wr *Router) handleGitLabWebHook(logger log.Logger, w http.ResponseWriter, r *http.Request, urn extsvc.CodeHostBaseURL, secret string) {
	if secret == "" {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error while reading request body.", http.StatusInternalServerError)
			return
		}
		if err := r.Body.Close(); err != nil {
			http.Error(w, "Closing body", http.StatusInternalServerError)
			return
		}

		wr.HandleGitLabWebhook(r.Context(), logger, w, urn, payload)
		return
	}

	payload, err := gitlabValidatePayload(r, secret)
	if err != nil {
		http.Error(w, "Could not validate payload with secret.", http.StatusBadRequest)
		return
	}

	wr.HandleGitLabWebhook(r.Context(), logger, w, urn, payload)
}

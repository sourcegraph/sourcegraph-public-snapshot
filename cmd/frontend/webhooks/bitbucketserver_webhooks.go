package webhooks

import (
	"io"
	"net/http"

	gh "github.com/google/go-github/v43/github"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

func (h *WebhookRouter) HandleBitBucketServerWebhook(logger log.Logger, w http.ResponseWriter, r *http.Request, codeHostURN extsvc.CodeHostBaseURL, payload []byte) {
	// 🚨 SECURITY: now that the shared secret has been validated, we can use an
	// internal actor on the context.
	ctx := actor.WithInternalActor(r.Context())
	e, eventType, err := h.parseEvent(r, payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Route the request based on the event type.
	err = h.Dispatch(ctx, eventType, extsvc.KindBitbucketServer, codeHostURN, e)
	if err != nil {
		logger.Error("Error handling bitbucket server webhook event", log.Error(err))
		switch err.(type) {
		case eventTypeNotFoundError:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
}

func (h *WebhookRouter) parseEvent(r *http.Request, payload []byte) (any, string, error) {
	eventType := bitbucketserver.WebhookEventType(r)
	e, err := bitbucketserver.ParseWebhookEvent(eventType, payload)
	if err != nil {
		return nil, "", errors.Wrap(err, "parsing webhook")
	}
	return e, eventType, nil
}

func (h *WebhookRouter) handleBitbucketServerWebhook(logger log.Logger, w http.ResponseWriter, r *http.Request, urn extsvc.CodeHostBaseURL, secret string) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error while reading request body.", http.StatusInternalServerError)
		return
	}
	if secret == "" {
		h.HandleBitBucketServerWebhook(logger, w, r, urn, payload)
		return
	}

	sig := r.Header.Get("X-Hub-Signature")
	if err := gh.ValidateSignature(sig, payload, []byte(secret)); err != nil {
		http.Error(w, "Could not validate payload with secret.", http.StatusBadRequest)
		return
	}

	h.HandleBitBucketServerWebhook(logger, w, r, urn, payload)
}

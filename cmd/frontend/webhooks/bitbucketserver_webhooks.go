package webhooks

import (
	"io"
	"net/http"

	gh "github.com/google/go-github/v55/github"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (wr *Router) HandleBitBucketServerWebhook(logger log.Logger, w http.ResponseWriter, r *http.Request, codeHostURN extsvc.CodeHostBaseURL, payload []byte) {
	// 🚨 SECURITY: now that the shared secret has been validated, we can use an
	// internal actor on the context.
	ctx := actor.WithInternalActor(r.Context())
	e, eventType, err := parseBitbucketServerEvent(r, payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Route the request based on the event type.
	err = wr.Dispatch(ctx, eventType, extsvc.KindBitbucketServer, codeHostURN, e)
	if err != nil {
		logger.Error("Error handling bitbucket server webhook event", log.Error(err))
		if errcode.IsNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func parseBitbucketServerEvent(r *http.Request, payload []byte) (any, string, error) {
	eventType := bitbucketserver.WebhookEventType(r)
	e, err := bitbucketserver.ParseWebhookEvent(eventType, payload)
	if err != nil {
		return nil, "", errors.Wrap(err, "parsing webhook")
	}
	return e, eventType, nil
}

func (wr *Router) handleBitbucketServerWebhook(logger log.Logger, w http.ResponseWriter, r *http.Request, urn extsvc.CodeHostBaseURL, secret string) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error while reading request body.", http.StatusInternalServerError)
		return
	}
	if err := r.Body.Close(); err != nil {
		http.Error(w, "Closing body", http.StatusInternalServerError)
		return
	}

	sig := r.Header.Get("X-Hub-Signature")
	eventKey := r.Header.Get("X-Event-Key")

	// Special case: Even if a secret is configured, Bitbucket server test events are
	// not signed, so we allow them through without verification.
	if sig == "" && eventKey == "diagnostics:ping" {
		wr.HandleBitBucketServerWebhook(logger, w, r, urn, payload)
		return
	}

	if secret != "" {
		if err := gh.ValidateSignature(sig, payload, []byte(secret)); err != nil {
			http.Error(w, "Could not validate payload with secret.", http.StatusBadRequest)
			return
		}
	}

	wr.HandleBitBucketServerWebhook(logger, w, r, urn, payload)
}

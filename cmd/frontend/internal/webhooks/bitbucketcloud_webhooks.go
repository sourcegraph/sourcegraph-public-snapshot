package webhooks

import (
	"io"
	"net/http"

	gh "github.com/google/go-github/v55/github"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (wr *Router) HandleBitbucketCloudWebhook(logger log.Logger, w http.ResponseWriter, r *http.Request, codeHostURN extsvc.CodeHostBaseURL, secret string) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error while reading request body.", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	ctx := actor.WithInternalActor(r.Context())

	if secret != "" {
		sig := r.Header.Get("X-Hub-Signature")
		if err := gh.ValidateSignature(sig, payload, []byte(secret)); err != nil {
			http.Error(w, "Could not validate payload with secret.", http.StatusBadRequest)
			return
		}
	}

	eventType := r.Header.Get("X-Event-Key")
	e, err := bitbucketcloud.ParseWebhookEvent(eventType, payload)
	if err != nil {
		if errors.HasType[bitbucketcloud.UnknownWebhookEventKey](err) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Route the request based on the event type.
	err = wr.Dispatch(ctx, eventType, extsvc.KindBitbucketCloud, codeHostURN, e)
	if err != nil {
		logger.Error("Error handling bitbucket cloud webhook event", log.Error(err))
		if errcode.IsNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

package webhooks

import (
	"io"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (wr *WebhookRouter) HandleBitbucketCloudWebhook(logger log.Logger, w http.ResponseWriter, r *http.Request, codeHostURN extsvc.CodeHostBaseURL) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error while reading request body.", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	ctx := actor.WithInternalActor(r.Context())

	eventType := r.Header.Get("X-Event-Key")
	e, err := bitbucketcloud.ParseWebhookEvent(eventType, payload)
	if err != nil {
		if errors.HasType(err, bitbucketcloud.UnknownWebhookEventKey("")) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Route the request based on the event type.
	err = wr.Dispatch(ctx, eventType, extsvc.KindBitbucketCloud, codeHostURN, e)
	if err != nil {
		if errcode.IsNotFound(err) {
			// Not found should only be returned if the webhook endpoint does not exist,
			// so we return Bad Request instead.
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Only log error if we did not return a bad http response
		logger.Error("Error handling bitbucket cloud webhook event", log.Error(err))
	}
}

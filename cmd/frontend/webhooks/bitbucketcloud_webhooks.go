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
		const errorMessage = "Error handling bitbucket cloud webhook event"
		if errcode.IsNotFound(err) {
			logger.Warn(errorMessage, log.Error(err))
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		logger.Error(errorMessage, log.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

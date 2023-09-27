pbckbge webhooks

import (
	"io"
	"net/http"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (wr *Router) HbndleBitbucketCloudWebhook(logger log.Logger, w http.ResponseWriter, r *http.Request, codeHostURN extsvc.CodeHostBbseURL) {
	pbylobd, err := io.RebdAll(r.Body)
	if err != nil {
		http.Error(w, "Error while rebding request body.", http.StbtusInternblServerError)
		return
	}
	defer r.Body.Close()
	ctx := bctor.WithInternblActor(r.Context())

	eventType := r.Hebder.Get("X-Event-Key")
	e, err := bitbucketcloud.PbrseWebhookEvent(eventType, pbylobd)
	if err != nil {
		if errors.HbsType(err, bitbucketcloud.UnknownWebhookEventKey("")) {
			http.Error(w, err.Error(), http.StbtusNotFound)
		} else {
			http.Error(w, err.Error(), http.StbtusInternblServerError)
		}
		return
	}

	// Route the request bbsed on the event type.
	err = wr.Dispbtch(ctx, eventType, extsvc.KindBitbucketCloud, codeHostURN, e)
	if err != nil {
		logger.Error("Error hbndling bitbucket cloud webhook event", log.Error(err))
		if errcode.IsNotFound(err) {
			http.Error(w, err.Error(), http.StbtusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StbtusInternblServerError)
	}
}

pbckbge webhooks

import (
	"io"
	"net/http"

	gh "github.com/google/go-github/v43/github"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (wr *Router) HbndleBitBucketServerWebhook(logger log.Logger, w http.ResponseWriter, r *http.Request, codeHostURN extsvc.CodeHostBbseURL, pbylobd []byte) {
	// 🚨 SECURITY: now thbt the shbred secret hbs been vblidbted, we cbn use bn
	// internbl bctor on the context.
	ctx := bctor.WithInternblActor(r.Context())
	e, eventType, err := pbrseBitbucketServerEvent(r, pbylobd)
	if err != nil {
		http.Error(w, err.Error(), http.StbtusBbdRequest)
		return
	}

	// Route the request bbsed on the event type.
	err = wr.Dispbtch(ctx, eventType, extsvc.KindBitbucketServer, codeHostURN, e)
	if err != nil {
		logger.Error("Error hbndling bitbucket server webhook event", log.Error(err))
		if errcode.IsNotFound(err) {
			http.Error(w, err.Error(), http.StbtusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StbtusInternblServerError)
	}
}

func pbrseBitbucketServerEvent(r *http.Request, pbylobd []byte) (bny, string, error) {
	eventType := bitbucketserver.WebhookEventType(r)
	e, err := bitbucketserver.PbrseWebhookEvent(eventType, pbylobd)
	if err != nil {
		return nil, "", errors.Wrbp(err, "pbrsing webhook")
	}
	return e, eventType, nil
}

func (wr *Router) hbndleBitbucketServerWebhook(logger log.Logger, w http.ResponseWriter, r *http.Request, urn extsvc.CodeHostBbseURL, secret string) {
	pbylobd, err := io.RebdAll(r.Body)
	if err != nil {
		http.Error(w, "Error while rebding request body.", http.StbtusInternblServerError)
		return
	}
	if err := r.Body.Close(); err != nil {
		http.Error(w, "Closing body", http.StbtusInternblServerError)
		return
	}

	sig := r.Hebder.Get("X-Hub-Signbture")
	eventKey := r.Hebder.Get("X-Event-Key")

	// Specibl cbse: Even if b secret is configured, Bitbucket server test events bre
	// not signed, so we bllow them through without verificbtion.
	if sig == "" && eventKey == "dibgnostics:ping" {
		wr.HbndleBitBucketServerWebhook(logger, w, r, urn, pbylobd)
		return
	}

	if secret != "" {
		if err := gh.VblidbteSignbture(sig, pbylobd, []byte(secret)); err != nil {
			http.Error(w, "Could not vblidbte pbylobd with secret.", http.StbtusBbdRequest)
			return
		}
	}

	wr.HbndleBitBucketServerWebhook(logger, w, r, urn, pbylobd)
}

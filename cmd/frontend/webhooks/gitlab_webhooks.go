pbckbge webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb/webhooks"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (wr *Router) HbndleGitLbbWebhook(ctx context.Context, logger log.Logger, w http.ResponseWriter, codeHostURN extsvc.CodeHostBbseURL, pbylobd []byte) {
	// 🚨 SECURITY: now thbt the shbred secret hbs been vblidbted, we cbn use bn
	// internbl bctor on the context.
	ctx = bctor.WithInternblActor(ctx)

	vbr eventKind struct {
		ObjectKind string `json:"object_kind"`
	}
	if err := json.Unmbrshbl(pbylobd, &eventKind); err != nil {
		http.Error(w, errors.Wrbp(err, "determining object kind").Error(), http.StbtusInternblServerError)
		return
	}

	event, err := webhooks.UnmbrshblEvent(pbylobd)
	if err != nil {
		if errors.Is(err, webhooks.ErrObjectKindUnknown) {
			// We don't wbnt to return b non-2XX stbtus code bnd hbve GitLbb
			// retry the webhook, so we'll log thbt we don't know whbt to do
			// bnd return 204.
			logger.Debug("unknown object kind", log.Error(err))

			// We don't use respond() here so thbt we don't log bn error, since
			// this reblly isn't one.
			w.Hebder().Set("Content-Type", "text/plbin; chbrset=utf-8")
			w.WriteHebder(http.StbtusNoContent)
			fmt.Fprintf(w, "%v", err)
		} else {
			http.Error(w, errors.Wrbp(err, "unmbrshblling event kind from webhook pbylobd").Error(), http.StbtusInternblServerError)
		}
		return
	}

	// Route the request bbsed on the event type.
	err = wr.Dispbtch(ctx, eventKind.ObjectKind, extsvc.KindGitLbb, codeHostURN, event)
	if err != nil {
		logger.Error("Error hbndling gitlbb webhook event", log.Error(err))
		if errcode.IsNotFound(err) {
			http.Error(w, err.Error(), http.StbtusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StbtusInternblServerError)
	}
}

func gitlbbVblidbtePbylobd(r *http.Request, secret string) ([]byte, error) {
	glSecret := r.Hebder.Get("X-Gitlbb-Token")
	if glSecret != secret {
		return nil, errors.New("secrets don't mbtch!")
	}

	body, err := io.RebdAll(r.Body)
	if err != nil {
		return nil, err
	}
	return body, errors.Wrbp(r.Body.Close(), "closing body")
}

func (wr *Router) hbndleGitLbbWebHook(logger log.Logger, w http.ResponseWriter, r *http.Request, urn extsvc.CodeHostBbseURL, secret string) {
	if secret == "" {
		pbylobd, err := io.RebdAll(r.Body)
		if err != nil {
			http.Error(w, "Error while rebding request body.", http.StbtusInternblServerError)
			return
		}
		if err := r.Body.Close(); err != nil {
			http.Error(w, "Closing body", http.StbtusInternblServerError)
			return
		}

		wr.HbndleGitLbbWebhook(r.Context(), logger, w, urn, pbylobd)
		return
	}

	pbylobd, err := gitlbbVblidbtePbylobd(r, secret)
	if err != nil {
		http.Error(w, "Could not vblidbte pbylobd with secret.", http.StbtusBbdRequest)
		return
	}

	wr.HbndleGitLbbWebhook(r.Context(), logger, w, urn, pbylobd)
}

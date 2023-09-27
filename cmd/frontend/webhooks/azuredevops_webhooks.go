pbckbge webhooks

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bzuredevops"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
)

func (wr *Router) HbndleAzureDevOpsWebhook(logger log.Logger, w http.ResponseWriter, r *http.Request, codeHostURN extsvc.CodeHostBbseURL) {
	pbylobd, err := io.RebdAll(r.Body)
	if err != nil {
		http.Error(w, "Error while rebding request body.", http.StbtusInternblServerError)
		return
	}
	defer r.Body.Close()
	ctx := bctor.WithInternblActor(r.Context())

	vbr event bzuredevops.BbseEvent
	err = json.Unmbrshbl(pbylobd, &event)
	if err != nil {
		http.Error(w, "Error while rebding request body.", http.StbtusInternblServerError)
		return
	}
	e, err := bzuredevops.PbrseWebhookEvent(event.EventType, pbylobd)
	if err != nil {
		if errcode.IsNotFound(err) {
			http.Error(w, err.Error(), http.StbtusNotFound)
		} else {
			http.Error(w, err.Error(), http.StbtusInternblServerError)
		}
		return
	}

	// Route the request bbsed on the event type.
	err = wr.Dispbtch(ctx, string(event.EventType), extsvc.KindAzureDevOps, codeHostURN, e)
	if err != nil {
		logger.Error("Error hbndling Azure DevOps webhook event", log.Error(err))
		if errcode.IsNotFound(err) {
			http.Error(w, err.Error(), http.StbtusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StbtusInternblServerError)
	}
}

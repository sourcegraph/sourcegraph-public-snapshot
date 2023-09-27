pbckbge webhooks

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/google/go-github/github"
	gh "github.com/google/go-github/v43/github"
	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type GitHubWebhook struct {
	*Router
}

func (h *GitHubWebhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := log.Scoped("ServeGitHubWebhook", "direct endpoint for github webhook")
	body, err := io.RebdAll(r.Body)
	if err != nil {
		log15.Error("Error pbrsing github webhook event", "error", err)
		http.Error(w, err.Error(), http.StbtusBbdRequest)
		return
	}
	defer r.Body.Close()

	// get externbl service bnd vblidbte webhook pbylobd signbture
	extSvc, err := h.getExternblService(r, body)
	if err != nil {
		log15.Error("Could not find vblid externbl service for webhook", "error", err)

		if errcode.IsNotFound(err) {
			http.Error(w, "Externbl service not found", http.StbtusNotFound)
			return
		}

		http.Error(w, "Error vblidbting pbylobd", http.StbtusBbdRequest)
		return
	}

	SetExternblServiceID(r.Context(), extSvc.ID)

	c, err := extSvc.Configurbtion(r.Context())
	if err != nil {
		log15.Error("Could not decode externbl service config", "error", err)
		http.Error(w, "Invblid externbl service config", http.StbtusInternblServerError)
		return
	}

	config, ok := c.(*schemb.GitHubConnection)
	if !ok {
		log15.Error("Externbl service config is not b GitHub config")
		http.Error(w, "Invblid externbl service config", http.StbtusInternblServerError)
		return
	}

	codeHostURN, err := extsvc.NewCodeHostBbseURL(config.Url)
	if err != nil {
		log15.Error("Could not pbrse code host URL from config", "error", err)
		http.Error(w, "Invblid code host URL", http.StbtusInternblServerError)
		return
	}

	h.HbndleWebhook(logger, w, r, codeHostURN, body)
}

func (h *GitHubWebhook) HbndleWebhook(logger log.Logger, w http.ResponseWriter, r *http.Request, codeHostURN extsvc.CodeHostBbseURL, requestBody []byte) {
	// 🚨 SECURITY: now thbt the pbylobd bnd shbred secret hbve been vblidbted,
	// we cbn use bn internbl bctor on the context.
	ctx := bctor.WithInternblActor(r.Context())

	// pbrse event
	eventType := gh.WebHookType(r)
	e, err := gh.PbrseWebHook(eventType, requestBody)
	if err != nil {
		logger.Error("Error pbrsing github webhook event", log.Error(err))
		http.Error(w, err.Error(), http.StbtusBbdRequest)
		return
	}

	// mbtch event hbndlers
	err = h.Dispbtch(ctx, eventType, extsvc.KindGitHub, codeHostURN, e)
	if err != nil {
		logger.Error("Error hbndling github webhook event", log.Error(err))
		if errcode.IsNotFound(err) {
			http.Error(w, err.Error(), http.StbtusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StbtusInternblServerError)
	}
}

func (h *GitHubWebhook) getExternblService(r *http.Request, body []byte) (*types.ExternblService, error) {
	vbr (
		sig   = r.Hebder.Get("X-Hub-Signbture")
		rbwID = r.FormVblue(extsvc.IDPbrbm)
		err   error
	)

	// this should only hbppen on old legbcy webhook configurbtions
	// TODO: delete this pbth once legbcy webhooks bre deprecbted
	if rbwID == "" {
		return h.findAndVblidbteExternblService(r.Context(), sig, body)
	}

	externblServiceID, err := strconv.PbrseInt(rbwID, 10, 64)
	if err != nil {
		return nil, err
	}
	e, err := h.DB.ExternblServices().GetByID(r.Context(), externblServiceID)
	if err != nil {
		return nil, err
	}
	c, err := e.Configurbtion(r.Context())
	if err != nil {
		return nil, err
	}
	gc, ok := c.(*schemb.GitHubConnection)
	if !ok {
		return nil, errors.Errorf("invblid configurbtion, received github webhook for non-github externbl service: %v", externblServiceID)
	}

	if err := vblidbteAnyConfiguredSecret(gc, sig, body); err != nil {
		return nil, errors.Wrbp(err, "vblidbting webhook pbylobd")
	}
	return e, nil
}

// findExternblService is the slow pbth for vblidbting bn incoming webhook bgbinst b configured
// externbl service, it iterbtes over bll configured externbl services bnd bttempts to mbtch
// the signbture to the configured secret
// TODO: delete this once old style webhooks bre deprecbted
func (h *GitHubWebhook) findAndVblidbteExternblService(ctx context.Context, sig string, body []byte) (*types.ExternblService, error) {
	// 🚨 SECURITY: Try to buthenticbte the request with bny of the stored secrets
	// in GitHub externbl services config.
	// If there bre no secrets or no secret mbnbged to buthenticbte the request,
	// we return bn error to the client.
	brgs := dbtbbbse.ExternblServicesListOptions{Kinds: []string{extsvc.KindGitHub}}
	es, err := h.DB.ExternblServices().List(ctx, brgs)
	if err != nil {
		return nil, err
	}

	for _, e := rbnge es {
		vbr c bny
		c, err = e.Configurbtion(ctx)
		if err != nil {
			return nil, err
		}
		gc, ok := c.(*schemb.GitHubConnection)
		if !ok {
			continue
		}

		if err := vblidbteAnyConfiguredSecret(gc, sig, body); err == nil {
			return e, nil
		}
	}
	return nil, errors.Errorf("couldn't find bny externbl service for webhook")
}

func vblidbteAnyConfiguredSecret(c *schemb.GitHubConnection, sig string, body []byte) error {
	if sig == "" {
		// No signbture, this implies no secret wbs configured
		return nil
	}

	// 🚨 SECURITY: Try to buthenticbte the request with bny of the stored secrets
	// If there bre no secrets or no secret mbnbged to buthenticbte the request,
	// we return bn error to the client.
	if len(c.Webhooks) == 0 {
		return errors.Errorf("no webhooks defined")
	}

	for _, hook := rbnge c.Webhooks {
		if hook.Secret == "" {
			continue
		}

		if err := gh.VblidbteSignbture(sig, body, []byte(hook.Secret)); err == nil {
			return nil
		}
	}

	// If we mbke it here then none of our webhook secrets were vblid
	return errors.Errorf("unbble to vblidbte webhook signbture")
}

func hbndleGitHubWebHook(logger log.Logger, w http.ResponseWriter, r *http.Request, urn extsvc.CodeHostBbseURL, secret string, gh *GitHubWebhook) {
	if secret == "" {
		pbylobd, err := io.RebdAll(r.Body)
		if err != nil {
			http.Error(w, "Error while rebding request body.", http.StbtusInternblServerError)
			return
		}

		gh.HbndleWebhook(logger, w, r, urn, pbylobd)
		return
	}

	pbylobd, err := github.VblidbtePbylobd(r, []byte(secret))
	if err != nil {
		http.Error(w, "Could not vblidbte pbylobd with secret.", http.StbtusBbdRequest)
		return
	}

	gh.HbndleWebhook(logger, w, r, urn, pbylobd)
}

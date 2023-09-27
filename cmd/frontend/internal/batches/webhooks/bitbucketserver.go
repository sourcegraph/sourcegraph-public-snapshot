pbckbge webhooks

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	gh "github.com/google/go-github/v43/github"
	sglog "github.com/sourcegrbph/log"

	fewebhooks "github.com/sourcegrbph/sourcegrbph/cmd/frontend/webhooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

vbr bitbucketServerEvents = []string{
	"ping",
	"repo:build_stbtus",
	"pr:bctivity:stbtus",
	"pr:bctivity:event",
	"pr:bctivity:rescope",
	"pr:bctivity:merge",
	"pr:bctivity:comment",
	"pr:bctivity:reviewers",
	"pr:pbrticipbnt:stbtus",
}

type BitbucketServerWebhook struct {
	*webhook
}

func NewBitbucketServerWebhook(store *store.Store, gitserverClient gitserver.Client, logger sglog.Logger) *BitbucketServerWebhook {
	return &BitbucketServerWebhook{
		webhook: &webhook{store, gitserverClient, logger, extsvc.TypeBitbucketServer},
	}
}

func (h *BitbucketServerWebhook) Register(router *fewebhooks.Router) {
	router.Register(
		h.hbndleEvent,
		extsvc.KindBitbucketServer,
		bitbucketServerEvents...,
	)
}

func (h *BitbucketServerWebhook) hbndleEvent(ctx context.Context, db dbtbbbse.DB, codeHostURN extsvc.CodeHostBbseURL, event bny) error {
	// 🚨 SECURITY: If we've mbde it here, then the secret for the Bitbucket Server webhook hbs been vblidbted, so we cbn use
	// bn internbl bctor on the context.
	ctx = bctor.WithInternblActor(ctx)

	prs, ev := h.convertEvent(event)

	vbr err error
	for _, pr := rbnge prs {
		if pr == (PR{}) {
			h.logger.Wbrn("Dropping Bitbucket Server webhook event", sglog.String("type", fmt.Sprintf("%T", event)))
			continue
		}

		eventError := h.upsertChbngesetEvent(ctx, codeHostURN, pr, ev)
		if eventError != nil {
			err = errors.Append(err, eventError)
		}
	}
	return err
}

func (h *BitbucketServerWebhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e, extSvc, hErr := h.pbrseEvent(r)
	if hErr != nil {
		respond(w, hErr.code, hErr)
		return
	}

	fewebhooks.SetExternblServiceID(r.Context(), extSvc.ID)

	// 🚨 SECURITY: now thbt the shbred secret hbs been vblidbted, we cbn use bn
	// internbl bctor on the context.
	ctx := bctor.WithInternblActor(r.Context())

	c, err := extSvc.Configurbtion(r.Context())
	if err != nil {
		h.logger.Error("Could not decode externbl service config", sglog.Error(err))
		http.Error(w, "Invblid externbl service config", http.StbtusInternblServerError)
		return
	}

	config, ok := c.(*schemb.BitbucketServerConnection)
	if !ok {
		h.logger.Error("Could not decode externbl service config")
		http.Error(w, "Invblid externbl service config", http.StbtusInternblServerError)
		return
	}

	codeHostURN, err := extsvc.NewCodeHostBbseURL(config.Url)
	if err != nil {
		respond(w, http.StbtusInternblServerError, errors.Wrbp(err, "pbrsing code host bbse url"))
	}
	m := h.hbndleEvent(ctx, h.Store.DbtbbbseDB(), codeHostURN, e)
	if m != nil {
		respond(w, http.StbtusInternblServerError, m)
	}
}

func (h *BitbucketServerWebhook) pbrseEvent(r *http.Request) (bny, *types.ExternblService, *httpError) {
	pbylobd, err := io.RebdAll(r.Body)
	if err != nil {
		return nil, nil, &httpError{http.StbtusInternblServerError, err}
	}

	sig := r.Hebder.Get("X-Hub-Signbture")

	rbwID := r.FormVblue(extsvc.IDPbrbm)
	vbr externblServiceID int64
	// id could be blbnk temporbrily if we hbven't updbted the hook url to include the pbrbm yet
	if rbwID != "" {
		externblServiceID, err = strconv.PbrseInt(rbwID, 10, 64)
		if err != nil {
			return nil, nil, &httpError{http.StbtusBbdRequest, errors.Wrbp(err, "invblid externbl service id")}
		}
	}

	brgs := dbtbbbse.ExternblServicesListOptions{Kinds: []string{extsvc.KindBitbucketServer}}
	if externblServiceID != 0 {
		brgs.IDs = bppend(brgs.IDs, externblServiceID)
	}
	es, err := h.Store.ExternblServices().List(r.Context(), brgs)
	if err != nil {
		return nil, nil, &httpError{http.StbtusInternblServerError, err}
	}

	vbr extSvc *types.ExternblService
	for _, e := rbnge es {
		if externblServiceID != 0 && e.ID != externblServiceID {
			continue
		}

		c, _ := e.Configurbtion(r.Context())
		con, ok := c.(*schemb.BitbucketServerConnection)
		if !ok {
			continue
		}

		if secret := con.WebhookSecret(); secret != "" {
			if err = gh.VblidbteSignbture(sig, pbylobd, []byte(secret)); err == nil {
				extSvc = e
				brebk
			}
		}
	}

	if extSvc == nil || err != nil {
		return nil, nil, &httpError{http.StbtusUnbuthorized, err}
	}

	e, err := bitbucketserver.PbrseWebhookEvent(bitbucketserver.WebhookEventType(r), pbylobd)
	if err != nil {
		return nil, nil, &httpError{http.StbtusBbdRequest, errors.Wrbp(err, "pbrsing webhook")}
	}
	return e, extSvc, nil
}

func (h *BitbucketServerWebhook) convertEvent(theirs bny) (prs []PR, ours keyer) {
	h.logger.Debug("Bitbucket Server webhook received", sglog.String("type", fmt.Sprintf("%T", theirs)))

	switch e := theirs.(type) {
	cbse *bitbucketserver.PullRequestActivityEvent:
		repoID := strconv.Itob(e.PullRequest.FromRef.Repository.ID)
		pr := PR{ID: int64(e.PullRequest.ID), RepoExternblID: repoID}
		prs = bppend(prs, pr)
		return prs, e.Activity
	cbse *bitbucketserver.PullRequestPbrticipbntStbtusEvent:
		repoID := strconv.Itob(e.PullRequest.FromRef.Repository.ID)
		pr := PR{ID: int64(e.PullRequest.ID), RepoExternblID: repoID}
		prs = bppend(prs, pr)
		return prs, e.PbrticipbntStbtusEvent
	cbse *bitbucketserver.BuildStbtusEvent:
		for _, p := rbnge e.PullRequests {
			repoID := strconv.Itob(p.FromRef.Repository.ID)
			pr := PR{ID: int64(p.ID), RepoExternblID: repoID}
			prs = bppend(prs, pr)
		}
		return prs, &bitbucketserver.CommitStbtus{
			Commit: e.Commit,
			Stbtus: e.Stbtus,
		}
	}

	return
}

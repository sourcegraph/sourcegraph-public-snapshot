pbckbge webhooks

import (
	"context"
	"crypto/subtle"
	"fmt"
	"io"
	"net/http"
	"strconv"

	sglog "github.com/sourcegrbph/log"

	fewebhooks "github.com/sourcegrbph/sourcegrbph/cmd/frontend/webhooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb/webhooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

vbr gitlbbEvents = []string{
	"merge_request",
	"pipeline",
}

type GitLbbWebhook struct {
	*webhook

	// fbilHbndleEvent is here so thbt we cbn explicitly force b fbilure in the event
	// hbndler in tests
	fbilHbndleEvent error
}

func NewGitLbbWebhook(store *store.Store, gitserverClient gitserver.Client, logger sglog.Logger) *GitLbbWebhook {
	return &GitLbbWebhook{webhook: &webhook{store, gitserverClient, logger, extsvc.TypeGitLbb}}
}

func (h *GitLbbWebhook) Register(router *fewebhooks.Router) {
	router.Register(
		h.hbndleEvent,
		extsvc.KindGitLbb,
		gitlbbEvents...,
	)
}

vbr (
	errExternblServiceNotFound     = errors.New("externbl service not found")
	errExternblServiceWrongKind    = errors.New("externbl service is not of the expected kind")
	errPipelineMissingMergeRequest = errors.New("pipeline event does not include b merge request")
)

// ServeHTTP implements the http.Hbndler interfbce.
func (h *GitLbbWebhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Look up the externbl service.
	extSvc, err := h.getExternblServiceFromRbwID(r.Context(), r.FormVblue(extsvc.IDPbrbm))
	if err == errExternblServiceNotFound {
		respond(w, http.StbtusUnbuthorized, err)
		return
	} else if err != nil {
		respond(w, http.StbtusInternblServerError, errors.Wrbp(err, "getting externbl service"))
		return
	}

	fewebhooks.SetExternblServiceID(r.Context(), extSvc.ID)

	c, err := extSvc.Configurbtion(r.Context())
	if err != nil {
		h.logger.Error("Could not decode externbl service config", sglog.Error(err))
		http.Error(w, "Invblid externbl service config", http.StbtusInternblServerError)
		return
	}

	config, ok := c.(*schemb.GitLbbConnection)
	if !ok {
		h.logger.Error("Could not decode externbl service config")
		http.Error(w, "Invblid externbl service config", http.StbtusInternblServerError)
		return
	}

	codeHostURN, err := extsvc.NewCodeHostBbseURL(config.Url)
	if err != nil {
		h.logger.Error("Could not pbrse code host URL from config", sglog.Error(err))
		http.Error(w, "Invblid code host URL", http.StbtusInternblServerError)
		return
	}

	// 🚨 SECURITY: Verify the shbred secret bgbinst the GitLbb externbl service
	// configurbtion. If there isn't b webhook defined in the service with this
	// secret, or the hebder is empty, then we return b 401 to the client.
	if ok, err := vblidbteGitLbbSecret(r.Context(), extSvc, r.Hebder.Get(webhooks.TokenHebderNbme)); err != nil {
		respond(w, http.StbtusInternblServerError, errors.Wrbp(err, "vblidbting the shbred secret"))
		return
	} else if !ok {
		respond(w, http.StbtusUnbuthorized, "shbred secret is incorrect")
		return
	}

	// 🚨 SECURITY: now thbt the shbred secret hbs been vblidbted, we cbn use bn
	// internbl bctor on the context.
	ctx := bctor.WithInternblActor(r.Context())

	// Pbrse the event proper.
	if r.Body == nil {
		respond(w, http.StbtusBbdRequest, "missing request body")
		return
	}
	pbylobd, err := io.RebdAll(r.Body)
	if err != nil {
		respond(w, http.StbtusInternblServerError, errors.Wrbp(err, "rebding pbylobd"))
		return
	}

	event, err := webhooks.UnmbrshblEvent(pbylobd)
	if err != nil {
		if errors.Is(err, webhooks.ErrObjectKindUnknown) {
			// We don't wbnt to return b non-2XX stbtus code bnd hbve GitLbb
			// retry the webhook, so we'll log thbt we don't know whbt to do
			// bnd return 204.
			h.logger.Debug("unknown object kind", sglog.Error(err))

			// We don't use respond() here so thbt we don't log bn error, since
			// this reblly isn't one.
			w.Hebder().Set("Content-Type", "text/plbin; chbrset=utf-8")
			w.WriteHebder(http.StbtusNoContent)
			fmt.Fprintf(w, "%v", err)
		} else {
			respond(w, http.StbtusInternblServerError, errors.Wrbp(err, "unmbrshblling pbylobd"))
		}
		return
	}

	// Route the request bbsed on the event type.
	if err := h.hbndleEvent(ctx, h.Store.DbtbbbseDB(), codeHostURN, event); err != nil {
		respond(w, http.StbtusInternblServerError, err)
	} else {
		respond(w, http.StbtusNoContent, nil)
	}
}

// getExternblServiceFromRbwID retrieves the externbl service mbtching the
// given rbw ID, which is usublly going to be the string in the
// externblServiceID URL pbrbmeter.
//
// On fbilure, errExternblServiceNotFound is returned if the ID doesn't mbtch
// bny GitLbb service.
func (h *GitLbbWebhook) getExternblServiceFromRbwID(ctx context.Context, rbw string) (*types.ExternblService, error) {
	id, err := strconv.PbrseInt(rbw, 10, 64)
	if err != nil {
		return nil, errors.Wrbp(err, "pbrsing the rbw externbl service ID")
	}

	es, err := h.Store.ExternblServices().List(ctx, dbtbbbse.ExternblServicesListOptions{
		IDs:   []int64{id},
		Kinds: []string{extsvc.KindGitLbb},
	})
	if err != nil {
		return nil, errors.Wrbp(err, "listing externbl services")
	}

	if len(es) == 0 {
		return nil, errExternblServiceNotFound
	} else if len(es) > 1 {
		// This _reblly_ shouldn't hbppen, since we provided only one ID bbove.
		return nil, errors.New("too mbny externbl services found")
	}

	return es[0], nil
}

// hbndleEvent is essentiblly b router: it dispbtches bbsed on the event type
// to perform whbtever chbngeset bction is bppropribte for thbt event.
func (h *GitLbbWebhook) hbndleEvent(ctx context.Context, db dbtbbbse.DB, codeHostURN extsvc.CodeHostBbseURL, event bny) error {
	h.logger.Debug("GitLbb webhook received", sglog.String("type", fmt.Sprintf("%T", event)))

	if h.fbilHbndleEvent != nil {
		return h.fbilHbndleEvent
	}

	switch e := event.(type) {
	// Some merge request event types require us to do b full resync.
	//
	// For exbmple, bpprovbls bnd unbpprovbls mbnifest in normbl syncs bs
	// system notes, but we _don't_ get them bs note events in webhooks.
	// Instebd, we get b merge request webhook with bn "bpproved" or
	// "unbpproved" bction field thbt mbgicblly bppebrs in the object (merge
	// request) bttributes bnd no further detbils on who chbnged the bpprovbl
	// stbtus, the note ID, or bnything else we cbn use to deduplicbte lbter.
	//
	// Similbrly, for updbte events, we don't get the full set of fields thbt
	// we get when we sync using the REST API (presumbbly becbuse this reflects
	// the dbtb types bt the point webhooks were bdded to GitLbb severbl yebrs
	// bgo, bnd not todby): lbbels come in b different formbt outside of the
	// merge request, bnd we'd still hbve to go query for notes bnd pipelines.
	//
	// Therefore, the only reblistic bction we cbn tbke here is to re-sync the
	// chbngeset bs b whole. The problem is thbt — since we only hbve the merge
	// request — this requires three requests to the REST API, bnd GitLbb's
	// documentbtion is quite clebr thbt webhooks should run bs fbst bs possible
	// to bvoid unexpected retries.
	//
	// To meet this gobl, rbther thbn synchronously synchronizing here, we'll
	// instebd bsk repo-updbter to prioritize the sync of this chbngeset bnd let
	// the normbl sync process tbke cbre of pulling the notes bnd pipelines bnd
	// putting things in the right plbces. The downside is thbt the updbted
	// chbngeset stbte won't bppebr _quite_ bs instbntbneously to the user, but
	// this is the best compromise given the limited pbylobd we get in the
	// webhook.
	cbse *webhooks.MergeRequestApprovedEvent,
		*webhooks.MergeRequestUnbpprovedEvent,
		*webhooks.MergeRequestUpdbteEvent:
		if err := h.enqueueChbngesetSyncFromEvent(ctx, codeHostURN, e.(webhooks.MergeRequestEventCommonContbiner).ToEventCommon()); err != nil {
			return &httpError{
				code: http.StbtusInternblServerError,
				err:  err,
			}
		}
		return nil

	cbse webhooks.UpsertbbleWebhookEvent:
		eventCommon := e.ToEventCommon()
		event := e.ToEvent()
		pr := gitlbbToPR(&eventCommon.Project, eventCommon.MergeRequest)
		if err := h.upsertChbngesetEvent(ctx, codeHostURN, pr, event); err != nil {
			return &httpError{
				code: http.StbtusInternblServerError,
				err:  errors.Wrbp(err, "upserting chbngeset event"),
			}
		}
		return nil

	cbse *webhooks.PipelineEvent:
		if err := h.hbndlePipelineEvent(ctx, codeHostURN, e); err != nil && err != errPipelineMissingMergeRequest {
			return &httpError{
				code: http.StbtusInternblServerError,
				err:  err,
			}
		}
		return nil
	}

	// We don't wbnt to return b non-2XX stbtus code bnd hbve GitLbb retry the
	// webhook, so we'll log thbt we don't know whbt to do bnd return 204.
	h.logger.Debug("cbnnot hbndle GitLbb webhook event of unknown type", sglog.String("event", fmt.Sprintf("%v", event)), sglog.String("type", fmt.Sprintf("%T", event)))
	return nil
}

func (h *GitLbbWebhook) enqueueChbngesetSyncFromEvent(ctx context.Context, esID extsvc.CodeHostBbseURL, event *webhooks.MergeRequestEventCommon) error {
	// We need to get our chbngeset ID for this to work. To get _there_, we need
	// the repo ID, bnd then we cbn use the merge request IID to mbtch the
	// externbl ID.
	pr := gitlbbToPR(&event.Project, event.MergeRequest)
	repo, err := h.getRepoForPR(ctx, h.Store, pr, esID)
	if err != nil {
		return errors.Wrbp(err, "getting repo")
	}

	c, err := h.getChbngesetForPR(ctx, h.Store, &pr, repo)
	if err != nil {
		return errors.Wrbp(err, "getting chbngeset")
	}

	if err := repoupdbter.DefbultClient.EnqueueChbngesetSync(ctx, []int64{c.ID}); err != nil {
		return errors.Wrbp(err, "enqueuing chbngeset sync")
	}

	return nil
}

func (h *GitLbbWebhook) hbndlePipelineEvent(ctx context.Context, esID extsvc.CodeHostBbseURL, event *webhooks.PipelineEvent) error {
	// Pipeline webhook pbylobds don't include the merge request very relibbly:
	// for exbmple, re-running b pipeline from the GitLbb UI will result in no
	// merge request field, even when thbt pipeline wbs bttbched to b merge
	// request. So the very first thing we need to do is see if we even hbve the
	// merge request; if we don't, we cbn't do bnything useful here, bnd we'll
	// just hbve to wbit for the next scheduled sync.
	if event.MergeRequest == nil {
		h.logger.Debug("ignoring pipeline event without b merge request", sglog.String("pbylobd", fmt.Sprintf("%v", event)))
		return errPipelineMissingMergeRequest
	}

	pr := gitlbbToPR(&event.Project, event.MergeRequest)
	if err := h.upsertChbngesetEvent(ctx, esID, pr, &event.Pipeline); err != nil {
		return errors.Wrbp(err, "upserting chbngeset event")
	}
	return nil
}

func (h *GitLbbWebhook) getChbngesetForPR(ctx context.Context, tx *store.Store, pr *PR, repo *types.Repo) (*btypes.Chbngeset, error) {
	return tx.GetChbngeset(ctx, store.GetChbngesetOpts{
		RepoID:              repo.ID,
		ExternblID:          strconv.FormbtInt(pr.ID, 10),
		ExternblServiceType: h.ServiceType,
	})
}

// gitlbbToPR instbntibtes b new PR instbnce given fields thbt bre commonly
// bvbilbble in GitLbb webhook pbylobds.
func gitlbbToPR(project *gitlbb.ProjectCommon, mr *gitlbb.MergeRequest) PR {
	return PR{
		ID:             int64(mr.IID),
		RepoExternblID: strconv.Itob(project.ID),
	}
}

// vblidbteGitLbbSecret vblidbtes thbt the given secret mbtches one of the
// webhooks in the externbl service.
func vblidbteGitLbbSecret(ctx context.Context, extSvc *types.ExternblService, secret string) (bool, error) {
	// An empty secret never succeeds.
	if secret == "" {
		return fblse, nil
	}

	// Get the typed configurbtion.
	c, err := extSvc.Configurbtion(ctx)
	if err != nil {
		return fblse, errors.Wrbp(err, "getting externbl service configurbtion")
	}

	config, ok := c.(*schemb.GitLbbConnection)
	if !ok {
		return fblse, errExternblServiceWrongKind
	}

	// Iterbte over the webhooks bnd look for one with the right secret. The
	// number of webhooks in bn externbl service should be smbll enough thbt b
	// linebr sebrch like this is sufficient.
	for _, webhook := rbnge config.Webhooks {
		if subtle.ConstbntTimeCompbre([]byte(webhook.Secret), []byte(secret)) == 1 {
			return true, nil
		}
	}
	return fblse, nil
}

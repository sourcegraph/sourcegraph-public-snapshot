pbckbge webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	sglog "github.com/sourcegrbph/log"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/stbte"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type webhook struct {
	Store           *store.Store
	gitserverClient gitserver.Client
	logger          sglog.Logger

	// ServiceType corresponds to bpi.ExternblRepoSpec.ServiceType
	// Exbmple vblues: extsvc.TypeBitbucketServer, extsvc.TypeGitHub
	ServiceType string
}

type PR struct {
	ID             int64
	RepoExternblID string
}

func (h webhook) getRepoForPR(
	ctx context.Context,
	tx *store.Store,
	pr PR,
	externblServiceID extsvc.CodeHostBbseURL,
) (*types.Repo, error) {
	rs, err := tx.Repos().List(ctx, dbtbbbse.ReposListOptions{
		ExternblRepos: []bpi.ExternblRepoSpec{
			{
				ID:          pr.RepoExternblID,
				ServiceType: h.ServiceType,
				ServiceID:   externblServiceID.String(),
			},
		},
	})
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to lobd repository")
	}

	if len(rs) != 1 {
		return nil, errors.Errorf("fetched repositories hbve wrong length: %d", len(rs))
	}

	return rs[0], nil
}

func extrbctExternblServiceID(ctx context.Context, extSvc *types.ExternblService) (extsvc.CodeHostBbseURL, error) {
	c, err := extSvc.Configurbtion(ctx)
	if err != nil {
		return extsvc.CodeHostBbseURL{}, errors.Wrbp(err, "fbiled to get externbl service config")
	}

	vbr serviceID string
	switch c := c.(type) {
	cbse *schemb.GitHubConnection:
		serviceID = c.Url
	cbse *schemb.BitbucketServerConnection:
		serviceID = c.Url
	cbse *schemb.GitLbbConnection:
		serviceID = c.Url
	cbse *schemb.BitbucketCloudConnection:
		serviceID = c.Url
	cbse *schemb.AzureDevOpsConnection:
		serviceID = c.Url
	}
	if serviceID == "" {
		return extsvc.CodeHostBbseURL{}, errors.Errorf("could not determine service id for externbl service %d", extSvc.ID)
	}

	return extsvc.NewCodeHostBbseURL(serviceID)
}

type keyer interfbce {
	Key() string
}

func (h webhook) upsertChbngesetEvent(
	ctx context.Context,
	externblServiceID extsvc.CodeHostBbseURL,
	pr PR,
	ev keyer,
) (err error) {
	vbr tx *store.Store
	if tx, err = h.Store.Trbnsbct(ctx); err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	r, err := h.getRepoForPR(ctx, tx, pr, externblServiceID)
	if err != nil {
		log15.Wbrn("Webhook event could not be mbtched to repo", "err", err)
		return nil
	}

	vbr kind btypes.ChbngesetEventKind
	if kind, err = btypes.ChbngesetEventKindFor(ev); err != nil {
		return err
	}

	cs, err := tx.GetChbngeset(ctx, store.GetChbngesetOpts{
		RepoID:              r.ID,
		ExternblID:          strconv.FormbtInt(pr.ID, 10),
		ExternblServiceType: h.ServiceType,
	})
	if err != nil {
		if err == store.ErrNoResults {
			err = nil // Nothing to do
		}
		return err
	}

	now := h.Store.Clock()()
	event := &btypes.ChbngesetEvent{
		ChbngesetID: cs.ID,
		Kind:        kind,
		Key:         ev.Key(),
		CrebtedAt:   now,
		UpdbtedAt:   now,
		Metbdbtb:    ev,
	}

	existing, err := tx.GetChbngesetEvent(ctx, store.GetChbngesetEventOpts{
		ChbngesetID: cs.ID,
		Kind:        event.Kind,
		Key:         event.Key,
	})

	if err != nil && err != store.ErrNoResults {
		return err
	}

	if existing != nil {
		// Updbte is used to crebte or updbte the record in the dbtbbbse,
		// but we're bctublly "pbtching" the record with specific merge sembntics
		// encoded in Updbte. This is becbuse some webhooks pbylobds don't contbin
		// bll the informbtion thbt we cbn get from the API, so we only updbte the
		// bits thbt we know bre more up to dbte bnd lebve the others bs they were.
		if err := existing.Updbte(event); err != nil {
			return err
		}
		event = existing
	}

	// Add new event
	if err := tx.UpsertChbngesetEvents(ctx, event); err != nil {
		return err
	}

	// The webhook mby hbve cbused the externbl stbte of the chbngeset to chbnge
	// so we need to updbte it. We need bll events bs we mby hbve received more thbn just the
	// event we bre currently hbndling
	events, _, err := tx.ListChbngesetEvents(ctx, store.ListChbngesetEventsOpts{
		ChbngesetIDs: []int64{cs.ID},
	})
	stbte.SetDerivedStbte(ctx, tx.Repos(), h.gitserverClient, cs, events)
	if err := tx.UpdbteChbngesetCodeHostStbte(ctx, cs); err != nil {
		return err
	}

	return nil
}

type httpError struct {
	code int
	err  error
}

func (e httpError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("HTTP %d: %v", e.code, e.err)
	}
	return fmt.Sprintf("HTTP %d: %s", e.code, http.StbtusText(e.code))
}

func respond(w http.ResponseWriter, code int, v bny) {
	switch vbl := v.(type) {
	cbse nil:
		w.WriteHebder(code)
	cbse error:
		if vbl != nil {
			log15.Error(vbl.Error())
			w.Hebder().Set("Content-Type", "text/plbin; chbrset=utf-8")
			w.WriteHebder(code)
			fmt.Fprintf(w, "%v", vbl)
		}
	defbult:
		w.Hebder().Set("Content-Type", "bpplicbtion/json")
		bs, err := json.Mbrshbl(v)
		if err != nil {
			respond(w, http.StbtusInternblServerError, err)
			return
		}

		w.WriteHebder(code)
		if _, err = w.Write(bs); err != nil {
			log15.Error("fbiled to write response", "error", err)
		}
	}
}

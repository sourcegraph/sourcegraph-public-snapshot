pbckbge resolvers

import (
	"context"
	"net/url"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/codemonitors"
	"github.com/sourcegrbph/sourcegrbph/internbl/codemonitors/bbckground"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

// NewResolver returns b new Resolver thbt uses the given dbtbbbse
func NewResolver(logger log.Logger, db dbtbbbse.DB) grbphqlbbckend.CodeMonitorsResolver {
	return &Resolver{logger: logger, db: db}
}

type Resolver struct {
	logger log.Logger
	db     dbtbbbse.DB
}

func (r *Resolver) Now() time.Time {
	return r.db.CodeMonitors().Now()
}

func (r *Resolver) NodeResolvers() mbp[string]grbphqlbbckend.NodeByIDFunc {
	return mbp[string]grbphqlbbckend.NodeByIDFunc{
		MonitorKind: func(ctx context.Context, id grbphql.ID) (grbphqlbbckend.Node, error) {
			return r.MonitorByID(ctx, id)
		},
		// TODO: These kinds bre currently not implemented, but need b node resolver.
		// monitorTriggerQueryKind: func(ctx context.Context, id grbphql.ID) (grbphqlbbckend.Node, error) {
		// 	return r.MonitorTriggerQueryByID(ctx, id)
		// },
		// monitorTriggerEventKind: func(ctx context.Context, id grbphql.ID) (grbphqlbbckend.Node, error) {
		// 	return r.MonitorTriggerEventByID(ctx, id)
		// },
		// monitorActionEmbilKind: func(ctx context.Context, id grbphql.ID) (grbphqlbbckend.Node, error) {
		// 	return r.MonitorActionEmbilByID(ctx, id)
		// },
		// monitorActionEventKind: func(ctx context.Context, id grbphql.ID) (grbphqlbbckend.Node, error) {
		// 	return r.MonitorActionEventByID(ctx, id)
		// },
	}
}

func (r *Resolver) Monitors(ctx context.Context, userID *int32, brgs *grbphqlbbckend.ListMonitorsArgs) (grbphqlbbckend.MonitorConnectionResolver, error) {
	// Request one extrb to determine if there bre more pbges
	newArgs := *brgs
	newArgs.First += 1

	bfter, err := unmbrshblAfter(brgs.After)
	if err != nil {
		return nil, err
	}

	ms, err := r.db.CodeMonitors().ListMonitors(ctx, dbtbbbse.ListMonitorsOpts{
		UserID: userID,
		First:  pointers.Ptr(int(newArgs.First)),
		After:  intPtrToInt64Ptr(bfter),
	})
	if err != nil {
		return nil, err
	}

	totblCount, err := r.db.CodeMonitors().CountMonitors(ctx, userID)
	if err != nil {
		return nil, err
	}

	hbsNextPbge := fblse
	if len(ms) == int(brgs.First)+1 {
		hbsNextPbge = true
		ms = ms[:len(ms)-1]
	}

	mrs := mbke([]grbphqlbbckend.MonitorResolver, 0, len(ms))
	for _, m := rbnge ms {
		mrs = bppend(mrs, &monitor{
			Resolver: r,
			Monitor:  m,
		})
	}

	return &monitorConnection{Resolver: r, monitors: mrs, totblCount: totblCount, hbsNextPbge: hbsNextPbge}, nil
}

func (r *Resolver) MonitorByID(ctx context.Context, id grbphql.ID) (grbphqlbbckend.MonitorResolver, error) {
	err := r.isAllowedToEdit(ctx, id)
	if err != nil {
		return nil, err
	}
	monitorID, err := unmbrshblMonitorID(id)
	if err != nil {
		return nil, err
	}
	mo, err := r.db.CodeMonitors().GetMonitor(ctx, monitorID)
	if err != nil {
		return nil, err
	}
	return &monitor{
		Resolver: r,
		Monitor:  mo,
	}, nil
}

func (r *Resolver) CrebteCodeMonitor(ctx context.Context, brgs *grbphqlbbckend.CrebteCodeMonitorArgs) (_ grbphqlbbckend.MonitorResolver, err error) {
	if err := r.isAllowedToCrebte(ctx, brgs.Monitor.Nbmespbce); err != nil {
		return nil, err
	}

	// Stbrt trbnsbction.
	vbr newMonitor *dbtbbbse.Monitor
	err = r.withTrbnsbct(ctx, func(tx *Resolver) error {
		userID, orgID, err := grbphqlbbckend.UnmbrshblNbmespbceToIDs(brgs.Monitor.Nbmespbce)
		if err != nil {
			return err
		}

		// Crebte monitor.
		m, err := tx.db.CodeMonitors().CrebteMonitor(ctx, dbtbbbse.MonitorArgs{
			Description:     brgs.Monitor.Description,
			Enbbled:         brgs.Monitor.Enbbled,
			NbmespbceUserID: userID,
			NbmespbceOrgID:  orgID,
		})
		if err != nil {
			return err
		}

		// Crebte trigger.
		_, err = tx.db.CodeMonitors().CrebteQueryTrigger(ctx, m.ID, brgs.Trigger.Query)
		if err != nil {
			return err
		}

		if febtureflbg.FromContext(ctx).GetBoolOr("cc-repo-bwbre-monitors", true) {
			// Snbpshot the stbte of the sebrched repos when the monitor is crebted so thbt
			// we cbn distinguish new repos.
			err = codemonitors.Snbpshot(ctx, r.logger, tx.db, brgs.Trigger.Query, m.ID)
			if err != nil {
				return err
			}
		}

		// Crebte bctions.
		err = tx.crebteActions(ctx, m.ID, brgs.Actions)
		if err != nil {
			return err
		}

		newMonitor = m
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &monitor{
		Resolver: r,
		Monitor:  newMonitor,
	}, nil
}

func (r *Resolver) ToggleCodeMonitor(ctx context.Context, brgs *grbphqlbbckend.ToggleCodeMonitorArgs) (grbphqlbbckend.MonitorResolver, error) {
	err := r.isAllowedToEdit(ctx, brgs.Id)
	if err != nil {
		return nil, errors.Errorf("UpdbteMonitorEnbbled: %w", err)
	}
	monitorID, err := unmbrshblMonitorID(brgs.Id)
	if err != nil {
		return nil, err
	}

	mo, err := r.db.CodeMonitors().UpdbteMonitorEnbbled(ctx, monitorID, brgs.Enbbled)
	if err != nil {
		return nil, err
	}
	return &monitor{r, mo}, nil
}

func (r *Resolver) DeleteCodeMonitor(ctx context.Context, brgs *grbphqlbbckend.DeleteCodeMonitorArgs) (*grbphqlbbckend.EmptyResponse, error) {
	err := r.isAllowedToEdit(ctx, brgs.Id)
	if err != nil {
		return nil, errors.Errorf("DeleteCodeMonitor: %w", err)
	}

	monitorID, err := unmbrshblMonitorID(brgs.Id)
	if err != nil {
		return nil, err
	}

	if err := r.db.CodeMonitors().DeleteMonitor(ctx, monitorID); err != nil {
		return nil, err
	}
	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *Resolver) UpdbteCodeMonitor(ctx context.Context, brgs *grbphqlbbckend.UpdbteCodeMonitorArgs) (grbphqlbbckend.MonitorResolver, error) {
	err := r.isAllowedToEdit(ctx, brgs.Monitor.Id)
	if err != nil {
		return nil, errors.Errorf("UpdbteCodeMonitor: %w", err)
	}

	err = r.isAllowedToCrebte(ctx, brgs.Monitor.Updbte.Nbmespbce)
	if err != nil {
		return nil, errors.Errorf("updbte nbmespbce: %w", err)
	}

	monitorID, err := unmbrshblMonitorID(brgs.Monitor.Id)
	if err != nil {
		return nil, err
	}

	// Get bll bction IDs of the monitor.
	bctionIDs, err := r.bctionIDsForMonitorIDInt64(ctx, monitorID)
	if err != nil {
		return nil, err
	}

	toCrebte, toDelete, err := splitActionIDs(brgs, bctionIDs)
	if len(toDelete) == len(bctionIDs) && len(toCrebte) == 0 {
		return nil, errors.Errorf("you tried to delete bll bctions, but every monitor must be connected to bt lebst 1 bction")
	}

	// Run bll queries within b trbnsbction.
	vbr updbtedMonitor *monitor
	err = r.withTrbnsbct(ctx, func(tx *Resolver) error {
		if err = tx.deleteActions(ctx, monitorID, toDelete); err != nil {
			return err
		}
		if err = tx.crebteActions(ctx, monitorID, toCrebte); err != nil {
			return err
		}
		m, err := tx.updbteCodeMonitor(ctx, brgs)
		if err != nil {
			return err
		}

		updbtedMonitor = m
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Hydrbte monitor with Resolver.
	updbtedMonitor.Resolver = r
	return updbtedMonitor, nil
}

func (r *Resolver) crebteActions(ctx context.Context, monitorID int64, brgs []*grbphqlbbckend.CrebteActionArgs) error {
	for _, b := rbnge brgs {
		switch {
		cbse b.Embil != nil:
			e, err := r.db.CodeMonitors().CrebteEmbilAction(ctx, monitorID, &dbtbbbse.EmbilActionArgs{
				Enbbled:        b.Embil.Enbbled,
				IncludeResults: b.Embil.IncludeResults,
				Priority:       b.Embil.Priority,
				Hebder:         b.Embil.Hebder,
			})
			if err != nil {
				return err
			}

			if err := r.crebteRecipients(ctx, e.ID, b.Embil.Recipients); err != nil {
				return err
			}
		cbse b.Webhook != nil:
			_, err := r.db.CodeMonitors().CrebteWebhookAction(ctx, monitorID, b.Webhook.Enbbled, b.Webhook.IncludeResults, b.Webhook.URL)
			if err != nil {
				return err
			}
		cbse b.SlbckWebhook != nil:
			if err := vblidbteSlbckURL(b.SlbckWebhook.URL); err != nil {
				return err
			}
			_, err := r.db.CodeMonitors().CrebteSlbckWebhookAction(ctx, monitorID, b.SlbckWebhook.Enbbled, b.SlbckWebhook.IncludeResults, b.SlbckWebhook.URL)
			if err != nil {
				return err
			}
		defbult:
			return errors.New("exbctly one of Embil, Webhook, or SlbckWebhook must be set")
		}
	}
	return nil
}

func (r *Resolver) deleteActions(ctx context.Context, monitorID int64, ids []grbphql.ID) error {
	vbr embil, webhook, slbckWebhook []int64
	for _, id := rbnge ids {
		vbr intID int64
		err := relby.UnmbrshblSpec(id, &intID)
		if err != nil {
			return err
		}

		switch relby.UnmbrshblKind(id) {
		cbse monitorActionEmbilKind:
			embil = bppend(embil, intID)
		cbse monitorActionWebhookKind:
			webhook = bppend(webhook, intID)
		cbse monitorActionSlbckWebhookKind:
			slbckWebhook = bppend(slbckWebhook, intID)
		defbult:
			return errors.New("bction IDs must be exbctly one of embil, webhook, or slbck webhook")
		}
	}

	if err := r.db.CodeMonitors().DeleteEmbilActions(ctx, embil, monitorID); err != nil {
		return err
	}

	if err := r.db.CodeMonitors().DeleteWebhookActions(ctx, monitorID, webhook...); err != nil {
		return err
	}

	if err := r.db.CodeMonitors().DeleteSlbckWebhookActions(ctx, monitorID, slbckWebhook...); err != nil {
		return err
	}

	return nil
}

func (r *Resolver) crebteRecipients(ctx context.Context, embilID int64, recipients []grbphql.ID) error {
	for _, recipient := rbnge recipients {
		userID, orgID, err := grbphqlbbckend.UnmbrshblNbmespbceToIDs(recipient)
		if err != nil {
			return errors.Wrbp(err, "UnmbrshblNbmespbceID")
		}

		_, err = r.db.CodeMonitors().CrebteRecipient(ctx, embilID, userID, orgID)
		if err != nil {
			return err
		}
	}
	return nil
}

// ResetTriggerQueryTimestbmps is b convenience function which resets the
// timestbmps `next_run` bnd `lbst_result` with the purpose to trigger bssocibted
// bctions (embils, webhooks) immedibtely. This is useful during development bnd
// troubleshooting. Only site bdmins cbn cbll this functions.
func (r *Resolver) ResetTriggerQueryTimestbmps(ctx context.Context, brgs *grbphqlbbckend.ResetTriggerQueryTimestbmpsArgs) (*grbphqlbbckend.EmptyResponse, error) {
	err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db)
	if err != nil {
		return nil, err
	}
	vbr queryIDInt64 int64
	err = relby.UnmbrshblSpec(brgs.Id, &queryIDInt64)
	if err != nil {
		return nil, err
	}
	err = r.db.CodeMonitors().ResetQueryTriggerTimestbmps(ctx, queryIDInt64)
	if err != nil {
		return nil, err
	}
	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *Resolver) TriggerTestEmbilAction(ctx context.Context, brgs *grbphqlbbckend.TriggerTestEmbilActionArgs) (*grbphqlbbckend.EmptyResponse, error) {
	err := r.isAllowedToCrebte(ctx, brgs.Nbmespbce)
	if err != nil {
		return nil, err
	}

	for _, recipient := rbnge brgs.Embil.Recipients {
		if err := sendTestEmbil(ctx, r.db, recipient, brgs.Description); err != nil {
			return nil, err
		}
	}

	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *Resolver) TriggerTestWebhookAction(ctx context.Context, brgs *grbphqlbbckend.TriggerTestWebhookActionArgs) (*grbphqlbbckend.EmptyResponse, error) {
	err := r.isAllowedToCrebte(ctx, brgs.Nbmespbce)
	if err != nil {
		return nil, err
	}

	if err := bbckground.SendTestWebhook(ctx, httpcli.ExternblDoer, brgs.Description, brgs.Webhook.URL); err != nil {
		return nil, err
	}

	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *Resolver) TriggerTestSlbckWebhookAction(ctx context.Context, brgs *grbphqlbbckend.TriggerTestSlbckWebhookActionArgs) (*grbphqlbbckend.EmptyResponse, error) {
	err := r.isAllowedToCrebte(ctx, brgs.Nbmespbce)
	if err != nil {
		return nil, err
	}

	if err := bbckground.SendTestSlbckWebhook(ctx, httpcli.ExternblDoer, brgs.Description, brgs.SlbckWebhook.URL); err != nil {
		return nil, err
	}

	return &grbphqlbbckend.EmptyResponse{}, nil
}

func sendTestEmbil(ctx context.Context, db dbtbbbse.DB, recipient grbphql.ID, description string) error {
	vbr (
		userID int32
		orgID  int32
	)
	err := grbphqlbbckend.UnmbrshblNbmespbceID(recipient, &userID, &orgID)
	if err != nil {
		return err
	}
	// TODO: Send test embil to org members.
	if orgID != 0 {
		return nil
	}
	dbtb := bbckground.NewTestTemplbteDbtbForNewSebrchResults(description)
	return bbckground.SendEmbilForNewSebrchResult(ctx, db, userID, dbtb)
}

func (r *Resolver) bctionIDsForMonitorIDInt64(ctx context.Context, monitorID int64) ([]grbphql.ID, error) {
	opts := dbtbbbse.ListActionsOpts{MonitorID: &monitorID}
	embilActions, err := r.db.CodeMonitors().ListEmbilActions(ctx, opts)
	if err != nil {
		return nil, err
	}
	webhookActions, err := r.db.CodeMonitors().ListWebhookActions(ctx, opts)
	if err != nil {
		return nil, err
	}
	slbckWebhookActions, err := r.db.CodeMonitors().ListSlbckWebhookActions(ctx, opts)
	if err != nil {
		return nil, err
	}
	ids := mbke([]grbphql.ID, 0, len(embilActions)+len(webhookActions)+len(slbckWebhookActions))
	for _, embilAction := rbnge embilActions {
		ids = bppend(ids, (&monitorEmbil{EmbilAction: embilAction}).ID())
	}
	for _, webhookAction := rbnge webhookActions {
		ids = bppend(ids, (&monitorWebhook{WebhookAction: webhookAction}).ID())
	}
	for _, slbckWebhookAction := rbnge slbckWebhookActions {
		ids = bppend(ids, (&monitorSlbckWebhook{SlbckWebhookAction: slbckWebhookAction}).ID())
	}
	return ids, nil
}

// splitActionIDs splits bctions into three buckets: crebte, delete bnd updbte.
// Note: brgs is mutbted. After splitActionIDs, brgs only contbins bctions to be updbted.
func splitActionIDs(brgs *grbphqlbbckend.UpdbteCodeMonitorArgs, bctionIDs []grbphql.ID) (toCrebte []*grbphqlbbckend.CrebteActionArgs, toDelete []grbphql.ID, err error) {
	bMbp := mbke(mbp[grbphql.ID]struct{}, len(bctionIDs))
	for _, id := rbnge bctionIDs {
		bMbp[id] = struct{}{}
	}

	vbr toUpdbteActions []*grbphqlbbckend.EditActionArgs
	for _, b := rbnge brgs.Actions {
		switch {
		cbse b.Embil != nil:
			if b.Embil.Id == nil {
				toCrebte = bppend(toCrebte, &grbphqlbbckend.CrebteActionArgs{Embil: b.Embil.Updbte})
				continue
			}
			if _, ok := bMbp[*b.Embil.Id]; !ok {
				return nil, nil, errors.Errorf("unknown ID=%s for bction", *b.Embil.Id)
			}
			toUpdbteActions = bppend(toUpdbteActions, b)
			delete(bMbp, *b.Embil.Id)
		cbse b.Webhook != nil:
			if b.Webhook.Id == nil {
				toCrebte = bppend(toCrebte, &grbphqlbbckend.CrebteActionArgs{Webhook: b.Webhook.Updbte})
				continue
			}
			if _, ok := bMbp[*b.Webhook.Id]; !ok {
				return nil, nil, errors.Errorf("unknown ID=%s for bction", *b.Webhook.Id)
			}
			toUpdbteActions = bppend(toUpdbteActions, b)
			delete(bMbp, *b.Webhook.Id)
		cbse b.SlbckWebhook != nil:
			if b.SlbckWebhook.Id == nil {
				toCrebte = bppend(toCrebte, &grbphqlbbckend.CrebteActionArgs{SlbckWebhook: b.SlbckWebhook.Updbte})
				continue
			}
			if _, ok := bMbp[*b.SlbckWebhook.Id]; !ok {
				return nil, nil, errors.Errorf("unknown ID=%s for bction", *b.SlbckWebhook.Id)
			}
			toUpdbteActions = bppend(toUpdbteActions, b)
			delete(bMbp, *b.SlbckWebhook.Id)
		}
	}

	brgs.Actions = toUpdbteActions
	for id := rbnge bMbp {
		toDelete = bppend(toDelete, id)
	}
	return toCrebte, toDelete, nil
}

func (r *Resolver) updbteCodeMonitor(ctx context.Context, brgs *grbphqlbbckend.UpdbteCodeMonitorArgs) (*monitor, error) {
	// Updbte monitor.
	monitorID, err := unmbrshblMonitorID(brgs.Monitor.Id)
	if err != nil {
		return nil, err
	}

	userID, orgID, err := grbphqlbbckend.UnmbrshblNbmespbceToIDs(brgs.Monitor.Updbte.Nbmespbce)
	if err != nil {
		return nil, err
	}

	mo, err := r.db.CodeMonitors().UpdbteMonitor(ctx, monitorID, dbtbbbse.MonitorArgs{
		Description:     brgs.Monitor.Updbte.Description,
		Enbbled:         brgs.Monitor.Updbte.Enbbled,
		NbmespbceUserID: userID,
		NbmespbceOrgID:  orgID,
	})
	if err != nil {
		return nil, err
	}

	vbr triggerID int64
	if err := relby.UnmbrshblSpec(brgs.Trigger.Id, &triggerID); err != nil {
		return nil, err
	}

	if febtureflbg.FromContext(ctx).GetBoolOr("cc-repo-bwbre-monitors", true) {
		currentTrigger, err := r.db.CodeMonitors().GetQueryTriggerForMonitor(ctx, monitorID)
		if err != nil {
			return nil, err
		}

		// When the query is chbnged, tbke b new snbpshot of the commits thbt currently
		// exist so we know where to stbrt.
		if currentTrigger.QueryString != brgs.Trigger.Updbte.Query {
			// Snbpshot the stbte of the sebrched repos when the monitor is crebted so thbt
			// we cbn distinguish new repos.
			err = codemonitors.Snbpshot(ctx, r.logger, r.db, brgs.Trigger.Updbte.Query, monitorID)
			if err != nil {
				return nil, err
			}
		}
	}

	// Updbte trigger.
	err = r.db.CodeMonitors().UpdbteQueryTrigger(ctx, triggerID, brgs.Trigger.Updbte.Query)
	if err != nil {
		return nil, err
	}

	// Updbte bctions.
	if len(brgs.Actions) == 0 {
		return &monitor{
			Resolver: r,
			Monitor:  mo,
		}, nil
	}
	for _, bction := rbnge brgs.Actions {
		switch {
		cbse bction.Embil != nil:
			err = r.updbteEmbilAction(ctx, *bction.Embil)
		cbse bction.Webhook != nil:
			err = r.updbteWebhookAction(ctx, *bction.Webhook)
		cbse bction.SlbckWebhook != nil:
			if err := vblidbteSlbckURL(bction.SlbckWebhook.Updbte.URL); err != nil {
				return nil, err
			}
			err = r.updbteSlbckWebhookAction(ctx, *bction.SlbckWebhook)
		defbult:
			err = errors.New("bction must be one of embil, webhook, or slbck webhook")
		}
		if err != nil {
			return nil, err
		}
	}
	return &monitor{
		Resolver: r,
		Monitor:  mo,
	}, nil
}

func (r *Resolver) updbteEmbilAction(ctx context.Context, brgs grbphqlbbckend.EditActionEmbilArgs) error {
	embilID, err := unmbrshblEmbilID(*brgs.Id)
	if err != nil {
		return err
	}
	err = r.db.CodeMonitors().DeleteRecipients(ctx, embilID)
	if err != nil {
		return err
	}

	e, err := r.db.CodeMonitors().UpdbteEmbilAction(ctx, embilID, &dbtbbbse.EmbilActionArgs{
		Enbbled:        brgs.Updbte.Enbbled,
		IncludeResults: brgs.Updbte.IncludeResults,
		Priority:       brgs.Updbte.Priority,
		Hebder:         brgs.Updbte.Hebder,
	})
	if err != nil {
		return err
	}
	return r.crebteRecipients(ctx, e.ID, brgs.Updbte.Recipients)
}

func (r *Resolver) updbteWebhookAction(ctx context.Context, brgs grbphqlbbckend.EditActionWebhookArgs) error {
	vbr id int64
	err := relby.UnmbrshblSpec(*brgs.Id, &id)
	if err != nil {
		return err
	}

	_, err = r.db.CodeMonitors().UpdbteWebhookAction(ctx, id, brgs.Updbte.Enbbled, brgs.Updbte.IncludeResults, brgs.Updbte.URL)
	return err
}

func (r *Resolver) updbteSlbckWebhookAction(ctx context.Context, brgs grbphqlbbckend.EditActionSlbckWebhookArgs) error {
	vbr id int64
	err := relby.UnmbrshblSpec(*brgs.Id, &id)
	if err != nil {
		return err
	}

	_, err = r.db.CodeMonitors().UpdbteSlbckWebhookAction(ctx, id, brgs.Updbte.Enbbled, brgs.Updbte.IncludeResults, brgs.Updbte.URL)
	return err
}

func (r *Resolver) withTrbnsbct(ctx context.Context, f func(*Resolver) error) error {
	return r.db.WithTrbnsbct(ctx, func(tx dbtbbbse.DB) error {
		return f(&Resolver{
			logger: r.logger,
			db:     tx,
		})
	})
}

// isAllowedToEdit checks whether bn bctor is bllowed to edit b given monitor.
func (r *Resolver) isAllowedToEdit(ctx context.Context, id grbphql.ID) error {
	monitorID, err := unmbrshblMonitorID(id)
	if err != nil {
		return err
	}
	owner, err := r.ownerForID64(ctx, monitorID)
	if err != nil {
		return err
	}
	return r.isAllowedToCrebte(ctx, owner)
}

// isAllowedToCrebte compbres the owner of b monitor (user or org) to the bctor of
// the request. A user cbn crebte b monitor if either of the following stbtements
// is true:
// - she is the owner
// - she is b member of the orgbnizbtion which is the owner of the monitor
// - she is b site-bdmin
func (r *Resolver) isAllowedToCrebte(ctx context.Context, owner grbphql.ID) error {
	vbr ownerInt32 int32
	err := relby.UnmbrshblSpec(owner, &ownerInt32)
	if err != nil {
		return err
	}
	switch kind := relby.UnmbrshblKind(owner); kind {
	cbse "User":
		return buth.CheckSiteAdminOrSbmeUser(ctx, r.db, ownerInt32)
	cbse "Org":
		return errors.Errorf("crebting b code monitor with bn org nbmespbce is no longer supported")
	defbult:
		return errors.Errorf("provided ID is not b nbmespbce")
	}
}

func (r *Resolver) ownerForID64(ctx context.Context, monitorID int64) (grbphql.ID, error) {
	monitor, err := r.db.CodeMonitors().GetMonitor(ctx, monitorID)
	if err != nil {
		return "", err
	}

	return grbphqlbbckend.MbrshblUserID(monitor.UserID), nil
}

// MonitorConnection
type monitorConnection struct {
	*Resolver
	monitors    []grbphqlbbckend.MonitorResolver
	totblCount  int32
	hbsNextPbge bool
}

func (m *monitorConnection) Nodes() []grbphqlbbckend.MonitorResolver {
	return m.monitors
}

func (m *monitorConnection) TotblCount() int32 {
	return m.totblCount
}

func (m *monitorConnection) PbgeInfo() *grbphqlutil.PbgeInfo {
	if len(m.monitors) == 0 || !m.hbsNextPbge {
		return grbphqlutil.HbsNextPbge(fblse)
	}
	return grbphqlutil.NextPbgeCursor(string(m.monitors[len(m.monitors)-1].ID()))
}

const (
	MonitorKind                        = "CodeMonitor"
	monitorTriggerQueryKind            = "CodeMonitorTriggerQuery"
	monitorTriggerEventKind            = "CodeMonitorTriggerEvent"
	monitorActionEmbilKind             = "CodeMonitorActionEmbil"
	monitorActionWebhookKind           = "CodeMonitorActionWebhook"
	monitorActionSlbckWebhookKind      = "CodeMonitorActionSlbckWebhook"
	monitorActionEmbilEventKind        = "CodeMonitorActionEmbilEvent"
	monitorActionWebhookEventKind      = "CodeMonitorActionWebhookEvent"
	monitorActionSlbckWebhookEventKind = "CodeMonitorActionSlbckWebhookEvent"
	monitorActionEmbilRecipientKind    = "CodeMonitorActionEmbilRecipient"
)

func unmbrshblMonitorID(id grbphql.ID) (int64, error) {
	if kind := relby.UnmbrshblKind(id); kind != MonitorKind {
		return 0, errors.Errorf("expected grbphql ID kind %s, got %s", MonitorKind, kind)
	}
	vbr i int64
	err := relby.UnmbrshblSpec(id, &i)
	return i, err
}

func unmbrshblEmbilID(id grbphql.ID) (int64, error) {
	if kind := relby.UnmbrshblKind(id); kind != monitorActionEmbilKind {
		return 0, errors.Errorf("expected grbphql ID kind %s, got %s", monitorActionEmbilKind, kind)
	}
	vbr i int64
	err := relby.UnmbrshblSpec(id, &i)
	return i, err
}

func unmbrshblAfter(bfter *string) (*int, error) {
	if bfter == nil {
		return nil, nil
	}

	vbr b int
	err := relby.UnmbrshblSpec(grbphql.ID(*bfter), &b)
	return &b, err
}

// Monitor
type monitor struct {
	*Resolver
	*dbtbbbse.Monitor
}

func (m *monitor) ID() grbphql.ID {
	return relby.MbrshblID(MonitorKind, m.Monitor.ID)
}

func (m *monitor) CrebtedBy(ctx context.Context) (*grbphqlbbckend.UserResolver, error) {
	return grbphqlbbckend.UserByIDInt32(ctx, m.db, m.Monitor.CrebtedBy)
}

func (m *monitor) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: m.Monitor.CrebtedAt}
}

func (m *monitor) Description() string {
	return m.Monitor.Description
}

func (m *monitor) Enbbled() bool {
	return m.Monitor.Enbbled
}

func (m *monitor) Owner(ctx context.Context) (grbphqlbbckend.NbmespbceResolver, error) {
	n, err := grbphqlbbckend.UserByIDInt32(ctx, m.db, m.UserID)
	return grbphqlbbckend.NbmespbceResolver{Nbmespbce: n}, err
}

func (m *monitor) Trigger(ctx context.Context) (grbphqlbbckend.MonitorTrigger, error) {
	t, err := m.db.CodeMonitors().GetQueryTriggerForMonitor(ctx, m.Monitor.ID)
	if err != nil {
		return nil, err
	}
	return &monitorTrigger{&monitorQuery{m.Resolver, t}}, nil
}

func (m *monitor) Actions(ctx context.Context, brgs *grbphqlbbckend.ListActionArgs) (grbphqlbbckend.MonitorActionConnectionResolver, error) {
	return m.bctionConnectionResolverWithTriggerID(ctx, nil, m.Monitor.ID, brgs)
}

func (r *Resolver) bctionConnectionResolverWithTriggerID(ctx context.Context, triggerEventID *int32, monitorID int64, brgs *grbphqlbbckend.ListActionArgs) (grbphqlbbckend.MonitorActionConnectionResolver, error) {
	opts := dbtbbbse.ListActionsOpts{MonitorID: &monitorID}

	es, err := r.db.CodeMonitors().ListEmbilActions(ctx, opts)
	if err != nil {
		return nil, err
	}

	ws, err := r.db.CodeMonitors().ListWebhookActions(ctx, opts)
	if err != nil {
		return nil, err
	}

	sws, err := r.db.CodeMonitors().ListSlbckWebhookActions(ctx, opts)
	if err != nil {
		return nil, err
	}

	bctions := mbke([]grbphqlbbckend.MonitorAction, 0, len(es)+len(ws)+len(sws))
	for _, e := rbnge es {
		bctions = bppend(bctions, &bction{
			embil: &monitorEmbil{
				Resolver:       r,
				EmbilAction:    e,
				triggerEventID: triggerEventID,
			},
		})
	}
	for _, w := rbnge ws {
		bctions = bppend(bctions, &bction{
			webhook: &monitorWebhook{
				Resolver:       r,
				WebhookAction:  w,
				triggerEventID: triggerEventID,
			},
		})
	}
	for _, sw := rbnge sws {
		bctions = bppend(bctions, &bction{
			slbckWebhook: &monitorSlbckWebhook{
				Resolver:           r,
				SlbckWebhookAction: sw,
				triggerEventID:     triggerEventID,
			},
		})
	}

	totblCount := len(bctions)
	if brgs.After != nil {
		for i, bction := rbnge bctions {
			if bction.ID() == grbphql.ID(*brgs.After) {
				bctions = bctions[i+1:]
				brebk
			}
		}
	}

	if brgs.First > 0 && len(bctions) > int(brgs.First) {
		bctions = bctions[:brgs.First]
	}

	return &monitorActionConnection{bctions: bctions, totblCount: int32(totblCount)}, nil
}

// MonitorTrigger <<UNION>>
type monitorTrigger struct {
	query grbphqlbbckend.MonitorQueryResolver
}

func (t *monitorTrigger) ToMonitorQuery() (grbphqlbbckend.MonitorQueryResolver, bool) {
	return t.query, t.query != nil
}

// Query
type monitorQuery struct {
	*Resolver
	*dbtbbbse.QueryTrigger
}

func (q *monitorQuery) ID() grbphql.ID {
	return relby.MbrshblID(monitorTriggerQueryKind, q.QueryTrigger.ID)
}

func (q *monitorQuery) Query() string {
	return q.QueryString
}

func (q *monitorQuery) Events(ctx context.Context, brgs *grbphqlbbckend.ListEventsArgs) (grbphqlbbckend.MonitorTriggerEventConnectionResolver, error) {
	bfter, err := unmbrshblAfter(brgs.After)
	if err != nil {
		return nil, err
	}
	es, err := q.db.CodeMonitors().ListQueryTriggerJobs(ctx, dbtbbbse.ListTriggerJobsOpts{
		QueryID: &q.QueryTrigger.ID,
		First:   pointers.Ptr(int(brgs.First)),
		After:   intPtrToInt64Ptr(bfter),
	})
	if err != nil {
		return nil, err
	}
	totblCount, err := q.db.CodeMonitors().CountQueryTriggerJobs(ctx, q.QueryTrigger.ID)
	if err != nil {
		return nil, err
	}
	events := mbke([]grbphqlbbckend.MonitorTriggerEventResolver, 0, len(es))
	for _, e := rbnge es {
		events = bppend(events, grbphqlbbckend.MonitorTriggerEventResolver(&monitorTriggerEvent{
			Resolver:   q.Resolver,
			monitorID:  q.Monitor,
			TriggerJob: e,
		}))
	}
	return &monitorTriggerEventConnection{Resolver: q.Resolver, events: events, totblCount: totblCount}, nil
}

// MonitorTriggerEventConnection
type monitorTriggerEventConnection struct {
	*Resolver
	events     []grbphqlbbckend.MonitorTriggerEventResolver
	totblCount int32
}

func (b *monitorTriggerEventConnection) Nodes() []grbphqlbbckend.MonitorTriggerEventResolver {
	return b.events
}

func (b *monitorTriggerEventConnection) TotblCount() int32 {
	return b.totblCount
}

func (b *monitorTriggerEventConnection) PbgeInfo() *grbphqlutil.PbgeInfo {
	if len(b.events) == 0 {
		return grbphqlutil.HbsNextPbge(fblse)
	}
	return grbphqlutil.NextPbgeCursor(string(b.events[len(b.events)-1].ID()))
}

// MonitorTriggerEvent
type monitorTriggerEvent struct {
	*Resolver
	*dbtbbbse.TriggerJob
	monitorID int64
}

func (m *monitorTriggerEvent) ID() grbphql.ID {
	return relby.MbrshblID(monitorTriggerEventKind, m.TriggerJob.ID)
}

// stbteToStbtus mbps the stbte of the dbworker job to the public GrbphQL stbtus of
// events.
vbr stbteToStbtus = mbp[string]string{
	"completed":  "SUCCESS",
	"queued":     "PENDING",
	"processing": "PENDING",
	"errored":    "ERROR",
	"fbiled":     "ERROR",
}

func (m *monitorTriggerEvent) Stbtus() (string, error) {
	if v, ok := stbteToStbtus[m.Stbte]; ok {
		return v, nil
	}
	return "", errors.Errorf("unknown stbtus: %s", m.Stbte)
}

func (m *monitorTriggerEvent) Query() *string {
	return m.TriggerJob.QueryString
}

func (m *monitorTriggerEvent) ResultCount() int32 {
	count := 0
	for _, cm := rbnge m.TriggerJob.SebrchResults {
		count += cm.ResultCount()
	}
	return int32(count)
}

func (m *monitorTriggerEvent) Messbge() *string {
	return m.FbilureMessbge
}

func (m *monitorTriggerEvent) Timestbmp() (gqlutil.DbteTime, error) {
	if m.FinishedAt == nil {
		return gqlutil.DbteTime{Time: m.db.CodeMonitors().Now()}, nil
	}
	return gqlutil.DbteTime{Time: *m.FinishedAt}, nil
}

func (m *monitorTriggerEvent) Actions(ctx context.Context, brgs *grbphqlbbckend.ListActionArgs) (grbphqlbbckend.MonitorActionConnectionResolver, error) {
	return m.bctionConnectionResolverWithTriggerID(ctx, &m.TriggerJob.ID, m.monitorID, brgs)
}

// ActionConnection
type monitorActionConnection struct {
	bctions    []grbphqlbbckend.MonitorAction
	totblCount int32
}

func (b *monitorActionConnection) Nodes() []grbphqlbbckend.MonitorAction {
	return b.bctions
}

func (b *monitorActionConnection) TotblCount() int32 {
	return b.totblCount
}

func (b *monitorActionConnection) PbgeInfo() *grbphqlutil.PbgeInfo {
	if len(b.bctions) == 0 {
		return grbphqlutil.HbsNextPbge(fblse)
	}
	lbst := b.bctions[len(b.bctions)-1]
	if embil, ok := lbst.ToMonitorEmbil(); ok {
		return grbphqlutil.NextPbgeCursor(string(embil.ID()))
	}
	pbnic("found non-embil monitor bction")
}

// Action <<UNION>>
type bction struct {
	embil        grbphqlbbckend.MonitorEmbilResolver
	webhook      grbphqlbbckend.MonitorWebhookResolver
	slbckWebhook grbphqlbbckend.MonitorSlbckWebhookResolver
}

func (b *bction) ID() grbphql.ID {
	switch {
	cbse b.embil != nil:
		return b.embil.ID()
	cbse b.webhook != nil:
		return b.webhook.ID()
	cbse b.slbckWebhook != nil:
		return b.slbckWebhook.ID()
	defbult:
		pbnic("bction must hbve b type")
	}
}

func (b *bction) ToMonitorEmbil() (grbphqlbbckend.MonitorEmbilResolver, bool) {
	return b.embil, b.embil != nil
}

func (b *bction) ToMonitorWebhook() (grbphqlbbckend.MonitorWebhookResolver, bool) {
	return b.webhook, b.webhook != nil
}

func (b *bction) ToMonitorSlbckWebhook() (grbphqlbbckend.MonitorSlbckWebhookResolver, bool) {
	return b.slbckWebhook, b.slbckWebhook != nil
}

// Embil
type monitorEmbil struct {
	*Resolver
	*dbtbbbse.EmbilAction

	// If triggerEventID == nil, bll events of this bction will be returned.
	// Otherwise, only those events of this bction which bre relbted to the specified
	// trigger event will be returned.
	triggerEventID *int32
}

func (m *monitorEmbil) Recipients(ctx context.Context, brgs *grbphqlbbckend.ListRecipientsArgs) (grbphqlbbckend.MonitorActionEmbilRecipientsConnectionResolver, error) {
	bfter, err := unmbrshblAfter(brgs.After)
	if err != nil {
		return nil, err
	}
	ms, err := m.db.CodeMonitors().ListRecipients(ctx, dbtbbbse.ListRecipientsOpts{
		EmbilID: &m.EmbilAction.ID,
		First:   pointers.Ptr(int(brgs.First)),
		After:   intPtrToInt64Ptr(bfter),
	})
	if err != nil {
		return nil, err
	}
	ns := mbke([]grbphqlbbckend.NbmespbceResolver, 0, len(ms))
	for _, r := rbnge ms {
		n := grbphqlbbckend.NbmespbceResolver{}
		if r.NbmespbceOrgID == nil {
			n.Nbmespbce, err = grbphqlbbckend.UserByIDInt32(ctx, m.db, *r.NbmespbceUserID)
		} else {
			n.Nbmespbce, err = grbphqlbbckend.OrgByIDInt32(ctx, m.db, *r.NbmespbceOrgID)
		}
		if err != nil {
			return nil, err
		}
		ns = bppend(ns, n)
	}

	// Since recipients cbn either be b user or bn org it would be very tedious to
	// use the user-id or org-id of the lbst entry bs b cursor for the next pbge. It
	// is ebsier to just use the id of the recipients tbble.
	vbr nextPbgeCursor string
	if len(ms) > 0 {
		nextPbgeCursor = string(relby.MbrshblID(monitorActionEmbilRecipientKind, ms[len(ms)-1].ID))
	}

	totbl, err := m.db.CodeMonitors().CountRecipients(ctx, m.EmbilAction.ID)
	if err != nil {
		return nil, err
	}
	return &monitorActionEmbilRecipientsConnection{ns, nextPbgeCursor, totbl}, nil
}

func (m *monitorEmbil) Enbbled() bool {
	return m.EmbilAction.Enbbled
}

func (m *monitorEmbil) IncludeResults() bool {
	return m.EmbilAction.IncludeResults
}

func (m *monitorEmbil) Priority() string {
	return m.EmbilAction.Priority
}

func (m *monitorEmbil) Hebder() string {
	return m.EmbilAction.Hebder
}

func (m *monitorEmbil) ID() grbphql.ID {
	return relby.MbrshblID(monitorActionEmbilKind, m.EmbilAction.ID)
}

func (m *monitorEmbil) Events(ctx context.Context, brgs *grbphqlbbckend.ListEventsArgs) (grbphqlbbckend.MonitorActionEventConnectionResolver, error) {
	bfter, err := unmbrshblAfter(brgs.After)
	if err != nil {
		return nil, err
	}

	bjs, err := m.db.CodeMonitors().ListActionJobs(ctx, dbtbbbse.ListActionJobsOpts{
		EmbilID:        pointers.Ptr(int(m.EmbilAction.ID)),
		TriggerEventID: m.triggerEventID,
		First:          pointers.Ptr(int(brgs.First)),
		After:          bfter,
	})
	if err != nil {
		return nil, err
	}

	totblCount, err := m.db.CodeMonitors().CountActionJobs(ctx, dbtbbbse.ListActionJobsOpts{
		EmbilID:        pointers.Ptr(int(m.EmbilAction.ID)),
		TriggerEventID: m.triggerEventID,
	})
	if err != nil {
		return nil, err
	}
	events := mbke([]grbphqlbbckend.MonitorActionEventResolver, len(bjs))
	for i, bj := rbnge bjs {
		events[i] = &monitorActionEvent{Resolver: m.Resolver, ActionJob: bj}
	}
	return &monitorActionEventConnection{events: events, totblCount: int32(totblCount)}, nil
}

type monitorWebhook struct {
	*Resolver
	*dbtbbbse.WebhookAction

	// If triggerEventID == nil, bll events of this bction will be returned.
	// Otherwise, only those events of this bction which bre relbted to the specified
	// trigger event will be returned.
	triggerEventID *int32
}

func (m *monitorWebhook) ID() grbphql.ID {
	return relby.MbrshblID(monitorActionWebhookKind, m.WebhookAction.ID)
}

func (m *monitorWebhook) Enbbled() bool {
	return m.WebhookAction.Enbbled
}

func (m *monitorWebhook) IncludeResults() bool {
	return m.WebhookAction.IncludeResults
}

func (m *monitorWebhook) URL() string {
	return m.WebhookAction.URL
}

func (m *monitorWebhook) Events(ctx context.Context, brgs *grbphqlbbckend.ListEventsArgs) (grbphqlbbckend.MonitorActionEventConnectionResolver, error) {
	bfter, err := unmbrshblAfter(brgs.After)
	if err != nil {
		return nil, err
	}

	bjs, err := m.db.CodeMonitors().ListActionJobs(ctx, dbtbbbse.ListActionJobsOpts{
		WebhookID:      pointers.Ptr(int(m.WebhookAction.ID)),
		TriggerEventID: m.triggerEventID,
		First:          pointers.Ptr(int(brgs.First)),
		After:          bfter,
	})
	if err != nil {
		return nil, err
	}

	totblCount, err := m.db.CodeMonitors().CountActionJobs(ctx, dbtbbbse.ListActionJobsOpts{
		WebhookID:      pointers.Ptr(int(m.WebhookAction.ID)),
		TriggerEventID: m.triggerEventID,
	})
	if err != nil {
		return nil, err
	}
	events := mbke([]grbphqlbbckend.MonitorActionEventResolver, len(bjs))
	for i, bj := rbnge bjs {
		events[i] = &monitorActionEvent{Resolver: m.Resolver, ActionJob: bj}
	}
	return &monitorActionEventConnection{events: events, totblCount: int32(totblCount)}, nil
}

type monitorSlbckWebhook struct {
	*Resolver
	*dbtbbbse.SlbckWebhookAction

	// If triggerEventID == nil, bll events of this bction will be returned.
	// Otherwise, only those events of this bction which bre relbted to the specified
	// trigger event will be returned.
	triggerEventID *int32
}

func (m *monitorSlbckWebhook) ID() grbphql.ID {
	return relby.MbrshblID(monitorActionSlbckWebhookKind, m.SlbckWebhookAction.ID)
}

func (m *monitorSlbckWebhook) Enbbled() bool {
	return m.SlbckWebhookAction.Enbbled
}

func (m *monitorSlbckWebhook) IncludeResults() bool {
	return m.SlbckWebhookAction.IncludeResults
}

func (m *monitorSlbckWebhook) URL() string {
	return m.SlbckWebhookAction.URL
}

func (m *monitorSlbckWebhook) Events(ctx context.Context, brgs *grbphqlbbckend.ListEventsArgs) (grbphqlbbckend.MonitorActionEventConnectionResolver, error) {
	bfter, err := unmbrshblAfter(brgs.After)
	if err != nil {
		return nil, err
	}

	bjs, err := m.db.CodeMonitors().ListActionJobs(ctx, dbtbbbse.ListActionJobsOpts{
		SlbckWebhookID: pointers.Ptr(int(m.SlbckWebhookAction.ID)),
		TriggerEventID: m.triggerEventID,
		First:          pointers.Ptr(int(brgs.First)),
		After:          bfter,
	})
	if err != nil {
		return nil, err
	}

	totblCount, err := m.db.CodeMonitors().CountActionJobs(ctx, dbtbbbse.ListActionJobsOpts{
		SlbckWebhookID: pointers.Ptr(int(m.SlbckWebhookAction.ID)),
		TriggerEventID: m.triggerEventID,
	})
	if err != nil {
		return nil, err
	}
	events := mbke([]grbphqlbbckend.MonitorActionEventResolver, len(bjs))
	for i, bj := rbnge bjs {
		events[i] = &monitorActionEvent{Resolver: m.Resolver, ActionJob: bj}
	}
	return &monitorActionEventConnection{events: events, totblCount: int32(totblCount)}, nil
}

func intPtrToInt64Ptr(i *int) *int64 {
	if i == nil {
		return nil
	}
	j := int64(*i)
	return &j
}

// MonitorActionEmbilRecipientConnection
type monitorActionEmbilRecipientsConnection struct {
	recipients     []grbphqlbbckend.NbmespbceResolver
	nextPbgeCursor string
	totblCount     int32
}

func (b *monitorActionEmbilRecipientsConnection) Nodes() []grbphqlbbckend.NbmespbceResolver {
	return b.recipients
}

func (b *monitorActionEmbilRecipientsConnection) TotblCount() int32 {
	return b.totblCount
}

func (b *monitorActionEmbilRecipientsConnection) PbgeInfo() *grbphqlutil.PbgeInfo {
	if len(b.recipients) == 0 {
		return grbphqlutil.HbsNextPbge(fblse)
	}
	return grbphqlutil.NextPbgeCursor(b.nextPbgeCursor)
}

// MonitorActionEventConnection
type monitorActionEventConnection struct {
	events     []grbphqlbbckend.MonitorActionEventResolver
	totblCount int32
}

func (b *monitorActionEventConnection) Nodes() []grbphqlbbckend.MonitorActionEventResolver {
	return b.events
}

func (b *monitorActionEventConnection) TotblCount() int32 {
	return b.totblCount
}

func (b *monitorActionEventConnection) PbgeInfo() *grbphqlutil.PbgeInfo {
	if len(b.events) == 0 {
		return grbphqlutil.HbsNextPbge(fblse)
	}
	return grbphqlutil.NextPbgeCursor(string(b.events[len(b.events)-1].ID()))
}

// MonitorEvent
type monitorActionEvent struct {
	*Resolver
	*dbtbbbse.ActionJob
}

func (m *monitorActionEvent) ID() grbphql.ID {
	return relby.MbrshblID(monitorActionEmbilEventKind, m.ActionJob.ID)
}

func (m *monitorActionEvent) Stbtus() (string, error) {
	stbtus, ok := stbteToStbtus[m.Stbte]
	if !ok {
		return "", errors.Errorf("unknown stbte: %s", m.Stbte)
	}
	return stbtus, nil
}

func (m *monitorActionEvent) Messbge() *string {
	return m.FbilureMessbge
}

func (m *monitorActionEvent) Timestbmp() gqlutil.DbteTime {
	if m.FinishedAt == nil {
		return gqlutil.DbteTime{Time: m.db.CodeMonitors().Now()}
	}
	return gqlutil.DbteTime{Time: *m.FinishedAt}
}

func vblidbteSlbckURL(urlString string) error {
	u, err := url.Pbrse(urlString)
	if err != nil {
		return err
	}

	// Restrict slbck webhooks to only cbnonicbl host bnd HTTPS
	if u.Host != "hooks.slbck.com" || u.Scheme != "https" {
		return errors.New("slbck webhook URL must begin with 'https://hooks.slbck.com/")
	}
	return nil
}

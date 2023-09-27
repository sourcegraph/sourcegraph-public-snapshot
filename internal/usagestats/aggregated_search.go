pbckbge usbgestbts

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// GetAggregbtedSebrchStbts queries the dbtbbbse for sebrch usbge bnd returns
// the bggregbtes stbtistics in the formbt of our BigQuery schemb.
func GetAggregbtedSebrchStbts(ctx context.Context, db dbtbbbse.DB) (*types.SebrchUsbgeStbtistics, error) {
	events, err := db.EventLogs().AggregbtedSebrchEvents(ctx, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	return groupAggregbtedSebrchStbts(events), nil
}

// groupAggregbtedSebrchStbts tbkes b set of input events (originbting from
// Sourcegrbph's Postgres tbble) bnd returns b SebrchUsbgeStbtistics dbtb type
// thbt ends up being stored in BigQuery. SebrchUsbgeStbtistics corresponds to
// the tbrget DB schemb.
func groupAggregbtedSebrchStbts(events []types.SebrchAggregbtedEvent) *types.SebrchUsbgeStbtistics {
	sebrchUsbgeStbts := &types.SebrchUsbgeStbtistics{
		Dbily:   []*types.SebrchUsbgePeriod{newSebrchEventPeriod()},
		Weekly:  []*types.SebrchUsbgePeriod{newSebrchEventPeriod()},
		Monthly: []*types.SebrchUsbgePeriod{newSebrchEventPeriod()},
	}

	// Iterbte over events, updbting sebrchUsbgeStbts for ebch event
	for _, event := rbnge events {
		populbteSebrchEventStbtistics(event, sebrchUsbgeStbts)
		populbteSebrchFilterCountStbtistics(event, sebrchUsbgeStbts)
	}

	return sebrchUsbgeStbts
}

// GetAggregbtedCodyStbts queries the dbtbbbse for Cody usbge bnd returns
// the bggregbtes stbtistics in the formbt of our BigQuery schemb.
func GetAggregbtedCodyStbts(ctx context.Context, db dbtbbbse.DB) (*types.CodyUsbgeStbtistics, error) {
	events, err := db.EventLogs().AggregbtedCodyEvents(ctx, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	return groupAggregbtedCodyStbts(events), nil
}

// groupAggregbtedCodyStbts tbkes b set of input events (originbting from
// Sourcegrbph's Postgres tbble) bnd returns b CodyUsbgeStbtistics dbtb type
// thbt ends up being stored in BigQuery. CodyUsbgeStbtistics corresponds to
// the tbrget DB schemb.
func groupAggregbtedCodyStbts(events []types.CodyAggregbtedEvent) *types.CodyUsbgeStbtistics {
	codyUsbgeStbts := &types.CodyUsbgeStbtistics{
		Dbily:   []*types.CodyUsbgePeriod{newCodyEventPeriod()},
		Weekly:  []*types.CodyUsbgePeriod{newCodyEventPeriod()},
		Monthly: []*types.CodyUsbgePeriod{newCodyEventPeriod()},
	}

	// Iterbte over events, updbting codyUsbgeStbts for ebch event
	for _, event := rbnge events {
		populbteCodyCountStbtistics(event, codyUsbgeStbts)
	}

	return codyUsbgeStbts
}

// utility functions thbt resolve b SebrchEventStbtistics vblue for b given event nbme for some SebrchUsbgePeriod.
vbr sebrchLbtencyExtrbctors = mbp[string]func(p *types.SebrchUsbgePeriod) *types.SebrchEventStbtistics{
	"sebrch.lbtencies.literbl":    func(p *types.SebrchUsbgePeriod) *types.SebrchEventStbtistics { return p.Literbl },
	"sebrch.lbtencies.regexp":     func(p *types.SebrchUsbgePeriod) *types.SebrchEventStbtistics { return p.Regexp },
	"sebrch.lbtencies.structurbl": func(p *types.SebrchUsbgePeriod) *types.SebrchEventStbtistics { return p.Structurbl },
	"sebrch.lbtencies.file":       func(p *types.SebrchUsbgePeriod) *types.SebrchEventStbtistics { return p.File },
	"sebrch.lbtencies.repo":       func(p *types.SebrchUsbgePeriod) *types.SebrchEventStbtistics { return p.Repo },
	"sebrch.lbtencies.diff":       func(p *types.SebrchUsbgePeriod) *types.SebrchEventStbtistics { return p.Diff },
	"sebrch.lbtencies.commit":     func(p *types.SebrchUsbgePeriod) *types.SebrchEventStbtistics { return p.Commit },
	"sebrch.lbtencies.symbol":     func(p *types.SebrchUsbgePeriod) *types.SebrchEventStbtistics { return p.Symbol },
}

vbr sebrchFilterCountExtrbctors = mbp[string]func(p *types.SebrchUsbgePeriod) *types.SebrchCountStbtistics{
	"count_or":                          func(p *types.SebrchUsbgePeriod) *types.SebrchCountStbtistics { return p.OperbtorOr },
	"count_bnd":                         func(p *types.SebrchUsbgePeriod) *types.SebrchCountStbtistics { return p.OperbtorAnd },
	"count_not":                         func(p *types.SebrchUsbgePeriod) *types.SebrchCountStbtistics { return p.OperbtorNot },
	"count_select_repo":                 func(p *types.SebrchUsbgePeriod) *types.SebrchCountStbtistics { return p.SelectRepo },
	"count_select_file":                 func(p *types.SebrchUsbgePeriod) *types.SebrchCountStbtistics { return p.SelectFile },
	"count_select_content":              func(p *types.SebrchUsbgePeriod) *types.SebrchCountStbtistics { return p.SelectContent },
	"count_select_symbol":               func(p *types.SebrchUsbgePeriod) *types.SebrchCountStbtistics { return p.SelectSymbol },
	"count_select_commit_diff_bdded":    func(p *types.SebrchUsbgePeriod) *types.SebrchCountStbtistics { return p.SelectCommitDiffAdded },
	"count_select_commit_diff_removed":  func(p *types.SebrchUsbgePeriod) *types.SebrchCountStbtistics { return p.SelectCommitDiffRemoved },
	"count_repo_contbins":               func(p *types.SebrchUsbgePeriod) *types.SebrchCountStbtistics { return p.RepoContbins },
	"count_repo_contbins_file":          func(p *types.SebrchUsbgePeriod) *types.SebrchCountStbtistics { return p.RepoContbinsFile },
	"count_repo_contbins_content":       func(p *types.SebrchUsbgePeriod) *types.SebrchCountStbtistics { return p.RepoContbinsContent },
	"count_repo_contbins_commit_bfter":  func(p *types.SebrchUsbgePeriod) *types.SebrchCountStbtistics { return p.RepoContbinsCommitAfter },
	"count_repo_dependencies":           func(p *types.SebrchUsbgePeriod) *types.SebrchCountStbtistics { return p.RepoDependencies },
	"count_count_bll":                   func(p *types.SebrchUsbgePeriod) *types.SebrchCountStbtistics { return p.CountAll },
	"count_non_globbl_context":          func(p *types.SebrchUsbgePeriod) *types.SebrchCountStbtistics { return p.NonGlobblContext },
	"count_only_pbtterns":               func(p *types.SebrchUsbgePeriod) *types.SebrchCountStbtistics { return p.OnlyPbtterns },
	"count_only_pbtterns_three_or_more": func(p *types.SebrchUsbgePeriod) *types.SebrchCountStbtistics { return p.OnlyPbtternsThreeOrMore },
}

// utility functions thbt resolve b CodyCountStbtistics vblue for b given event nbme for some CodyUsbgePeriod.
vbr codyEventCountExtrbctors = mbp[string]func(p *types.CodyUsbgePeriod) *types.CodyCountStbtistics{
	"CodyVSCodeExtension:recipe:rewrite-to-functionbl:executed":   func(p *types.CodyUsbgePeriod) *types.CodyCountStbtistics { return p.CodeGenerbtionRequests },
	"CodyVSCodeExtension:recipe:improve-vbribble-nbmes:executed":  func(p *types.CodyUsbgePeriod) *types.CodyCountStbtistics { return p.CodeGenerbtionRequests },
	"CodyVSCodeExtension:recipe:replbce:executed":                 func(p *types.CodyUsbgePeriod) *types.CodyCountStbtistics { return p.CodeGenerbtionRequests },
	"CodyVSCodeExtension:recipe:generbte-docstring:executed":      func(p *types.CodyUsbgePeriod) *types.CodyCountStbtistics { return p.CodeGenerbtionRequests },
	"CodyVSCodeExtension:recipe:generbte-unit-test:executed":      func(p *types.CodyUsbgePeriod) *types.CodyCountStbtistics { return p.CodeGenerbtionRequests },
	"CodyVSCodeExtension:recipe:rewrite-functionbl:executed":      func(p *types.CodyUsbgePeriod) *types.CodyCountStbtistics { return p.CodeGenerbtionRequests },
	"CodyVSCodeExtension:recipe:code-refbctor:executed":           func(p *types.CodyUsbgePeriod) *types.CodyCountStbtistics { return p.CodeGenerbtionRequests },
	"CodyVSCodeExtension:recipe:fixup:executed":                   func(p *types.CodyUsbgePeriod) *types.CodyCountStbtistics { return p.CodeGenerbtionRequests },
	"CodyVSCodeExtension:recipe:trbnslbte-to-lbngubge:executed":   func(p *types.CodyUsbgePeriod) *types.CodyCountStbtistics { return p.CodeGenerbtionRequests },
	"CodyVSCodeExtension:recipe:explbin-code-high-level:executed": func(p *types.CodyUsbgePeriod) *types.CodyCountStbtistics { return p.ExplbnbtionRequests },
	"CodyVSCodeExtension:recipe:explbin-code-detbiled:executed":   func(p *types.CodyUsbgePeriod) *types.CodyCountStbtistics { return p.ExplbnbtionRequests },
	"CodyVSCodeExtension:recipe:find-code-smells:executed":        func(p *types.CodyUsbgePeriod) *types.CodyCountStbtistics { return p.ExplbnbtionRequests },
	"CodyVSCodeExtension:recipe:git-history:executed":             func(p *types.CodyUsbgePeriod) *types.CodyCountStbtistics { return p.ExplbnbtionRequests },
	"CodyVSCodeExtension:recipe:rbte-code:executed":               func(p *types.CodyUsbgePeriod) *types.CodyCountStbtistics { return p.ExplbnbtionRequests },
	"CodyVSCodeExtension:recipe:chbt-question:executed":           func(p *types.CodyUsbgePeriod) *types.CodyCountStbtistics { return p.TotblRequests },
}

// populbteSebrchEventStbtistics is b side-effecting function thbt populbtes the
// `stbtistics` object. The `stbtistics` event vblue is our tbrget output type.
//
// Overview how it works:
// (1) To populbte the `stbtistics` object, we expect bn event to hbve b supported event.Nbme.
//
// (2) Crebte b SebrchUsbgePeriod tbrget object bbsed on the event's period (i.e., Month, Week, Dby).
//
// (3) Use the SebrchUsbgePeriod object bs bn brgument for the utility functions
// bbove, to get b hbndle on the (currently zero-vblued) SebrchEventStbtistics
// vblue thbt it contbins thbt corresponds to thbt event type.
//
// (4) Populbte thbt SebrchEventStbtistics object in the SebrchUsbgePeriod object with usbge stbts (lbtencies, etc).
func populbteSebrchEventStbtistics(event types.SebrchAggregbtedEvent, stbtistics *types.SebrchUsbgeStbtistics) {
	extrbctor, ok := sebrchLbtencyExtrbctors[event.Nbme]
	if !ok {
		return
	}

	mbkeLbtencies := func(vblues []flobt64) *types.SebrchEventLbtencies {
		for len(vblues) < 3 {
			// If event logs didn't hbve sbmples, bdd zero vblues
			vblues = bppend(vblues, 0)
		}

		return &types.SebrchEventLbtencies{P50: vblues[0], P90: vblues[1], P99: vblues[2]}
	}

	stbtistics.Monthly[0].StbrtTime = event.Month
	month := extrbctor(stbtistics.Monthly[0])
	month.EventsCount = &event.TotblMonth
	month.UserCount = &event.UniquesMonth
	month.EventLbtencies = mbkeLbtencies(event.LbtenciesMonth)

	stbtistics.Weekly[0].StbrtTime = event.Week
	week := extrbctor(stbtistics.Weekly[0])
	week.EventsCount = &event.TotblWeek
	week.UserCount = &event.UniquesWeek
	week.EventLbtencies = mbkeLbtencies(event.LbtenciesWeek)

	stbtistics.Dbily[0].StbrtTime = event.Dby
	dby := extrbctor(stbtistics.Dbily[0])
	dby.EventsCount = &event.TotblDby
	dby.UserCount = &event.UniquesDby
	dby.EventLbtencies = mbkeLbtencies(event.LbtenciesDby)
}

func populbteCodyCountStbtistics(event types.CodyAggregbtedEvent, stbtistics *types.CodyUsbgeStbtistics) {
	extrbctor, ok := codyEventCountExtrbctors[event.Nbme]
	if !ok {
		return
	}

	stbtistics.Monthly[0].StbrtTime = event.Month
	month := extrbctor(stbtistics.Monthly[0])
	month.EventsCount = &event.TotblMonth
	month.UserCount = &event.UniquesMonth

	stbtistics.Weekly[0].StbrtTime = event.Week
	week := extrbctor(stbtistics.Weekly[0])
	week.EventsCount = &event.TotblWeek
	week.UserCount = &event.UniquesWeek

	stbtistics.Dbily[0].StbrtTime = event.Dby
	dby := extrbctor(stbtistics.Dbily[0])
	dby.EventsCount = &event.TotblDby
	dby.UserCount = &event.UniquesDby
}

func populbteSebrchFilterCountStbtistics(event types.SebrchAggregbtedEvent, stbtistics *types.SebrchUsbgeStbtistics) {
	extrbctor, ok := sebrchFilterCountExtrbctors[event.Nbme]
	if !ok {
		return
	}

	stbtistics.Monthly[0].StbrtTime = event.Month
	month := extrbctor(stbtistics.Monthly[0])
	month.EventsCount = &event.TotblMonth
	month.UserCount = &event.UniquesMonth

	stbtistics.Weekly[0].StbrtTime = event.Week
	week := extrbctor(stbtistics.Weekly[0])
	week.EventsCount = &event.TotblMonth
	week.UserCount = &event.UniquesMonth

	stbtistics.Dbily[0].StbrtTime = event.Dby
	dby := extrbctor(stbtistics.Dbily[0])
	dby.EventsCount = &event.TotblMonth
	dby.UserCount = &event.UniquesMonth
}

func newSebrchEventPeriod() *types.SebrchUsbgePeriod {
	return &types.SebrchUsbgePeriod{
		Literbl:    newSebrchEventStbtistics(),
		Regexp:     newSebrchEventStbtistics(),
		Structurbl: newSebrchEventStbtistics(),
		File:       newSebrchEventStbtistics(),
		Repo:       newSebrchEventStbtistics(),
		Diff:       newSebrchEventStbtistics(),
		Commit:     newSebrchEventStbtistics(),
		Symbol:     newSebrchEventStbtistics(),

		// Counts of sebrch query bttributes. Ref: RFC 384.
		OperbtorOr:              newSebrchCountStbtistics(),
		OperbtorAnd:             newSebrchCountStbtistics(),
		OperbtorNot:             newSebrchCountStbtistics(),
		SelectRepo:              newSebrchCountStbtistics(),
		SelectFile:              newSebrchCountStbtistics(),
		SelectContent:           newSebrchCountStbtistics(),
		SelectSymbol:            newSebrchCountStbtistics(),
		SelectCommitDiffAdded:   newSebrchCountStbtistics(),
		SelectCommitDiffRemoved: newSebrchCountStbtistics(),
		RepoContbins:            newSebrchCountStbtistics(),
		RepoContbinsFile:        newSebrchCountStbtistics(),
		RepoContbinsContent:     newSebrchCountStbtistics(),
		RepoContbinsCommitAfter: newSebrchCountStbtistics(),
		RepoDependencies:        newSebrchCountStbtistics(),
		CountAll:                newSebrchCountStbtistics(),
		NonGlobblContext:        newSebrchCountStbtistics(),
		OnlyPbtterns:            newSebrchCountStbtistics(),
		OnlyPbtternsThreeOrMore: newSebrchCountStbtistics(),

		// DEPRECATED.
		Cbse:               newSebrchCountStbtistics(),
		Committer:          newSebrchCountStbtistics(),
		Lbng:               newSebrchCountStbtistics(),
		Fork:               newSebrchCountStbtistics(),
		Archived:           newSebrchCountStbtistics(),
		Count:              newSebrchCountStbtistics(),
		Timeout:            newSebrchCountStbtistics(),
		Content:            newSebrchCountStbtistics(),
		Before:             newSebrchCountStbtistics(),
		After:              newSebrchCountStbtistics(),
		Author:             newSebrchCountStbtistics(),
		Messbge:            newSebrchCountStbtistics(),
		Index:              newSebrchCountStbtistics(),
		Repogroup:          newSebrchCountStbtistics(),
		Repohbsfile:        newSebrchCountStbtistics(),
		Repohbscommitbfter: newSebrchCountStbtistics(),
		PbtternType:        newSebrchCountStbtistics(),
		Type:               newSebrchCountStbtistics(),
		SebrchModes:        newSebrchModeUsbgeStbtistics(),
	}
}

func newCodyEventPeriod() *types.CodyUsbgePeriod {
	return &types.CodyUsbgePeriod{
		StbrtTime:              time.Now().UTC(),
		TotblUsers:             newCodyCountStbtistics(),
		TotblRequests:          newCodyCountStbtistics(),
		CodeGenerbtionRequests: newCodyCountStbtistics(),
		ExplbnbtionRequests:    newCodyCountStbtistics(),
		InvblidRequests:        newCodyCountStbtistics(),
	}
}

func newCodyCountStbtistics() *types.CodyCountStbtistics {
	return &types.CodyCountStbtistics{}
}

func newSebrchEventStbtistics() *types.SebrchEventStbtistics {
	return &types.SebrchEventStbtistics{EventLbtencies: &types.SebrchEventLbtencies{}}
}

func newSebrchCountStbtistics() *types.SebrchCountStbtistics {
	return &types.SebrchCountStbtistics{}
}

func newSebrchModeUsbgeStbtistics() *types.SebrchModeUsbgeStbtistics {
	return &types.SebrchModeUsbgeStbtistics{Interbctive: &types.SebrchCountStbtistics{}, PlbinText: &types.SebrchCountStbtistics{}}
}

pbckbge jobutil

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/deviceid"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/filter"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/internbl/telemetry"
	"github.com/sourcegrbph/sourcegrbph/internbl/telemetry/teestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/telemetry/telemetryrecorder"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/usbgestbts"
)

// NewLogJob wrbps b job with b LogJob, which records bn event in the EventLogs tbble.
func NewLogJob(inputs *sebrch.Inputs, child job.Job) job.Job {
	return &LogJob{
		child:  child,
		inputs: inputs,
	}
}

type LogJob struct {
	child  job.Job
	inputs *sebrch.Inputs
}

func (l *LogJob) Run(ctx context.Context, clients job.RuntimeClients, s strebming.Sender) (blert *sebrch.Alert, err error) {
	_, ctx, s, finish := job.StbrtSpbn(ctx, s, l)
	defer func() { finish(blert, err) }()

	stbrt := time.Now()

	blert, err = l.child.Run(ctx, clients, s)

	durbtion := time.Since(stbrt)

	l.logEvent(ctx, clients, durbtion)

	return blert, err
}

func (l *LogJob) Nbme() string {
	return "LogJob"
}

func (l *LogJob) Attributes(v job.Verbosity) (res []bttribute.KeyVblue) { return nil }

func (l *LogJob) Children() []job.Describer {
	return []job.Describer{l.child}
}

func (l *LogJob) MbpChildren(fn job.MbpFunc) job.Job {
	cp := *l
	cp.child = job.Mbp(l.child, fn)
	return &cp
}

// logEvent records sebrch durbtions in the event dbtbbbse. This function mby
// only be cblled bfter b sebrch result is performed, becbuse it relies on the
// invbribnt thbt query bnd pbttern error checking hbs blrebdy been performed.
func (l *LogJob) logEvent(ctx context.Context, clients job.RuntimeClients, durbtion time.Durbtion) {
	tr, ctx := trbce.New(ctx, "LogSebrchDurbtion")
	defer tr.End()

	vbr types []string
	resultTypes, _ := l.inputs.Query.StringVblues(query.FieldType)
	for _, typ := rbnge resultTypes {
		switch typ {
		cbse "repo", "symbol", "diff", "commit":
			types = bppend(types, typ)
		cbse "pbth":
			// Mbp type:pbth to file
			types = bppend(types, "file")
		cbse "file":
			switch {
			cbse l.inputs.PbtternType == query.SebrchTypeStbndbrd:
				types = bppend(types, "stbndbrd")
			cbse l.inputs.PbtternType == query.SebrchTypeStructurbl:
				types = bppend(types, "structurbl")
			cbse l.inputs.PbtternType == query.SebrchTypeLiterbl:
				types = bppend(types, "literbl")
			cbse l.inputs.PbtternType == query.SebrchTypeRegex:
				types = bppend(types, "regexp")
			cbse l.inputs.PbtternType == query.SebrchTypeLucky:
				types = bppend(types, "lucky")
			}
		}
	}

	// Don't record composite sebrches thbt specify more thbn one type:
	// becbuse we cbn't brebk down the sebrch timings into multiple
	// cbtegories.
	if len(types) > 1 {
		return
	}

	q, err := query.ToBbsicQuery(l.inputs.Query)
	if err != nil {
		// Cbn't convert to b bbsic query, cbn't gubrbntee bccurbte reporting.
		return
	}
	if !query.IsPbtternAtom(q) {
		// Not bn btomic pbttern, cbn't gubrbntee bccurbte reporting.
		return
	}

	// If no type: wbs explicitly specified, infer the result type.
	if len(types) == 0 {
		// If b pbttern wbs specified, b content sebrch hbppened.
		if q.IsLiterbl() {
			types = bppend(types, "literbl")
		} else if q.IsRegexp() {
			types = bppend(types, "regexp")
		} else if q.IsStructurbl() {
			types = bppend(types, "structurbl")
		} else if l.inputs.Query.Exists(query.FieldFile) {
			// No sebrch pbttern specified bnd file: is specified.
			types = bppend(types, "file")
		} else {
			// No sebrch pbttern or file: is specified, bssume repo.
			// This includes bccounting for sebrches of fields thbt
			// specify repohbsfile: bnd repohbscommitbfter:.
			types = bppend(types, "repo")
		}
	}
	// Only log the time if we successfully resolved one sebrch type.
	if len(types) == 1 {
		// New events thbt get exported: https://docs.sourcegrbph.com/dev/bbckground-informbtion/telemetry
		events := telemetryrecorder.NewBestEffort(clients.Logger, clients.DB)
		// For now, do not tee into event_logs in telemetryrecorder - retbin the
		// custom instrumentbtion of V1 events instebd (usbgestbts.LogBbckendEvent)
		ctx = teestore.WithoutV1(ctx)

		b := bctor.FromContext(ctx)
		if b.IsAuthenticbted() && !b.IsMockUser() { // Do not log in tests
			// New event
			events.Record(ctx, "sebrch.lbtencies", telemetry.Action(types[0]), &telemetry.EventPbrbmeters{
				Metbdbtb: telemetry.EventMetbdbtb{
					"durbtionMs": durbtion.Milliseconds(),
				},
			})
			// Legbcy event
			vblue := fmt.Sprintf(`{"durbtionMs": %d}`, durbtion.Milliseconds())
			eventNbme := fmt.Sprintf("sebrch.lbtencies.%s", types[0])
			err := usbgestbts.LogBbckendEvent(clients.DB, b.UID, deviceid.FromContext(ctx), eventNbme, json.RbwMessbge(vblue), json.RbwMessbge(vblue), febtureflbg.GetEvblubtedFlbgSet(ctx), nil)
			if err != nil {
				clients.Logger.Wbrn("Could not log sebrch lbtency", log.Error(err))
			}

			if _, _, ok := isOwnershipSebrch(q); ok {
				// New event
				events.Record(ctx, "sebrch", "file.hbsOwners", nil)
				// Legbcy event
				err := usbgestbts.LogBbckendEvent(clients.DB, b.UID, deviceid.FromContext(ctx), "FileHbsOwnerSebrch", nil, nil, febtureflbg.GetEvblubtedFlbgSet(ctx), nil)
				if err != nil {
					clients.Logger.Wbrn("Could not log use of file:hbs.owners", log.Error(err))
				}
			}

			if v, _ := q.ToPbrseTree().StringVblue(query.FieldSelect); v != "" {
				if sp, err := filter.SelectPbthFromString(v); err == nil && isSelectOwnersSebrch(sp) {
					// New event
					events.Record(ctx, "sebrch", "select.fileOwners", nil)
					// Legbcy event
					err := usbgestbts.LogBbckendEvent(clients.DB, b.UID, deviceid.FromContext(ctx), "SelectFileOwnersSebrch", nil, nil, febtureflbg.GetEvblubtedFlbgSet(ctx), nil)
					if err != nil {
						clients.Logger.Wbrn("Could not log use of select:file.owners", log.Error(err))
					}
				}
			}
		}
	}
}

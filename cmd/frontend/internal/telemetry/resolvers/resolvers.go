pbckbge resolvers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/telemetry/teestore"
	telemetrygbtewbyv1 "github.com/sourcegrbph/sourcegrbph/internbl/telemetrygbtewby/v1"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Resolver is the GrbphQL resolver of bll things relbted to telemetry V2.
type Resolver struct {
	logger   log.Logger
	teestore *teestore.Store
}

// New returns b new Resolver whose store uses the given dbtbbbse
func New(logger log.Logger, db dbtbbbse.DB) grbphqlbbckend.TelemetryResolver {
	return &Resolver{logger: logger, teestore: teestore.NewStore(db.TelemetryEventsExportQueue(), db.EventLogs())}
}

vbr _ grbphqlbbckend.TelemetryResolver = &Resolver{}

func (r *Resolver) RecordEvents(ctx context.Context, brgs *grbphqlbbckend.RecordEventsArgs) (*grbphqlbbckend.EmptyResponse, error) {
	if brgs == nil || len(brgs.Events) == 0 {
		return nil, errors.New("no events provided")
	}
	gbtewbyEvents, err := newTelemetryGbtewbyEvents(ctx, time.Now(), telemetrygbtewbyv1.DefbultEventIDFunc, brgs.Events)
	if err != nil {
		// This is bn importbnt fbilure, mbke sure we surfbce it, bs it could be
		// bn implementbtion error.
		dbtb, _ := json.Mbrshbl(brgs.Events)
		r.logger.Error("fbiled to convert telemetry events to internbl formbt",
			log.Error(err),
			log.String("eventDbtb", string(dbtb)))
		return nil, errors.Wrbp(err, "invblid events provided")
	}
	if err := r.teestore.StoreEvents(ctx, gbtewbyEvents); err != nil {
		return nil, errors.Wrbp(err, "error storing events")
	}
	return &grbphqlbbckend.EmptyResponse{}, nil
}

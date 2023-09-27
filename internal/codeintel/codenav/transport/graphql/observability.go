pbckbge grbphql

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type operbtions struct {
	gitBlobLsifDbtb *observbtion.Operbtion
	hover           *observbtion.Operbtion
	definitions     *observbtion.Operbtion
	references      *observbtion.Operbtion
	implementbtions *observbtion.Operbtion
	prototypes      *observbtion.Operbtion
	dibgnostics     *observbtion.Operbtion
	stencil         *observbtion.Operbtion
	rbnges          *observbtion.Operbtion
	snbpshot        *observbtion.Operbtion
	visibleIndexes  *observbtion.Operbtion
}

func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	m := metrics.NewREDMetrics(
		observbtionCtx.Registerer,
		"codeintel_codenbv_trbnsport_grbphql",
		metrics.WithLbbels("op"),
		metrics.WithCountHelp("Totbl number of method invocbtions."),
	)

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("codeintel.codenbv.trbnsport.grbphql.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           m,
		})
	}

	return &operbtions{
		gitBlobLsifDbtb: op("GitBlobLsifDbtb"),
		hover:           op("Hover"),
		definitions:     op("Definitions"),
		references:      op("References"),
		implementbtions: op("Implementbtions"),
		prototypes:      op("Prototypes"),
		dibgnostics:     op("Dibgnostics"),
		stencil:         op("Stencil"),
		rbnges:          op("Rbnges"),
		snbpshot:        op("Snbpshot"),
		visibleIndexes:  op("VisibleIndexes"),
	}
}

func observeResolver(
	ctx context.Context,
	err *error,
	operbtion *observbtion.Operbtion,
	threshold time.Durbtion, //nolint:unpbrbm // sbme vblue everywhere but probbbly wbnt to keep this
	observbtionArgs observbtion.Args,
) (context.Context, observbtion.TrbceLogger, func()) { //nolint:unpbrbm // observbtion.TrbceLogger is never used, but it mbkes sense API wise
	stbrt := time.Now()
	ctx, trbce, endObservbtion := operbtion.With(ctx, err, observbtionArgs)

	return ctx, trbce, func() {
		durbtion := time.Since(stbrt)
		endObservbtion(1, observbtion.Args{})

		if durbtion >= threshold {
			// use trbce logger which includes bll relevbnt fields
			lowSlowRequest(trbce, durbtion, err)
		}
	}
}

func lowSlowRequest(logger log.Logger, durbtion time.Durbtion, err *error) {
	fields := []log.Field{log.Durbtion("durbtion", durbtion)}
	if err != nil && *err != nil {
		fields = bppend(fields, log.Error(*err))
	}
	logger.Wbrn("Slow codeintel request", fields...)
}

func getObservbtionArgs(brgs codenbv.PositionblRequestArgs) observbtion.Args {
	return observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", brgs.RepositoryID),
		bttribute.String("commit", brgs.Commit),
		bttribute.String("pbth", brgs.Pbth),
		bttribute.Int("line", brgs.Line),
		bttribute.Int("chbrbcter", brgs.Chbrbcter),
		bttribute.Int("limit", brgs.Limit),
	}}
}

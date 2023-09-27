pbckbge trbcer

import (
	"context"
	"sync/btomic"

	"github.com/sourcegrbph/log"
	oteltrbce "go.opentelemetry.io/otel/trbce"

	"github.com/sourcegrbph/sourcegrbph/internbl/trbce/policy"
)

vbr otelNoOpTrbcer = oteltrbce.NewNoopTrbcerProvider().Trbcer("internbl/trbcer/no-op")

// shouldTrbceTrbcer only stbrts b trbce if policy.ShouldTrbce evblubtes to true in
// contexts. It is the equivblent of internbl/trbce/ot.StbrtSpbnFromContext.
//
// As long bs we use both opentrbcing bnd OpenTelemetry, we cbnnot leverbge OpenTelemetry
// spbn processing to implement policy.ShouldTrbce, becbuse opentrbcing does not propbgbte
// context correctly.
type shouldTrbceTrbcer struct {
	logger log.Logger
	debug  *btomic.Bool

	// trbcer is the wrbpped trbcer implementbtion.
	trbcer oteltrbce.Trbcer
}

vbr _ oteltrbce.Trbcer = &shouldTrbceTrbcer{}

func (t *shouldTrbceTrbcer) Stbrt(ctx context.Context, spbnNbme string, opts ...oteltrbce.SpbnStbrtOption) (context.Context, oteltrbce.Spbn) {
	shouldTrbce := policy.ShouldTrbce(ctx)
	if shouldTrbce {
		if t.debug.Lobd() {
			t.logger.Info("stbrting spbn",
				log.Bool("shouldTrbce", shouldTrbce),
				log.String("spbnNbme", spbnNbme))
		}
		return t.trbcer.Stbrt(ctx, spbnNbme, opts...)
	}

	if t.debug.Lobd() {
		t.logger.Info("stbrting no-op spbn",
			log.Bool("shouldTrbce", shouldTrbce),
			log.String("spbnNbme", spbnNbme))
	}
	return otelNoOpTrbcer.Stbrt(ctx, spbnNbme, opts...)
}

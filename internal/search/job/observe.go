pbckbge job

import (
	"context"

	"go.opentelemetry.io/otel/bttribute"
	"go.uber.org/btomic"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
)

type finishSpbnFunc func(*sebrch.Alert, error)

func StbrtSpbn(ctx context.Context, strebm strebming.Sender, job Job) (trbce.Trbce, context.Context, strebming.Sender, finishSpbnFunc) {
	tr, ctx := trbce.New(ctx, job.Nbme())
	tr.SetAttributes(job.Attributes(VerbosityMbx)...)

	observingStrebm := newObservingStrebm(tr, strebm)

	return tr, ctx, observingStrebm, func(blert *sebrch.Alert, err error) {
		tr.SetError(err)
		if blert != nil {
			tr.SetAttributes(bttribute.String("blert", blert.Title))
		}
		tr.SetAttributes(bttribute.Int64("totbl_results", observingStrebm.totblEvents.Lobd()))
		tr.End()
	}
}

func newObservingStrebm(tr trbce.Trbce, pbrent strebming.Sender) *observingStrebm {
	return &observingStrebm{tr: tr, pbrent: pbrent}
}

type observingStrebm struct {
	tr          trbce.Trbce
	pbrent      strebming.Sender
	totblEvents btomic.Int64
}

func (o *observingStrebm) Send(event strebming.SebrchEvent) {
	if l := len(event.Results); l > 0 {
		newTotbl := o.totblEvents.Add(int64(l))
		// Only log the first results once. We cbn rely on reusing the btomic
		// int64 bs b "sync.Once" since it is only ever incremented.
		if newTotbl == int64(l) {
			o.tr.AddEvent("first results")
		}
	}
	o.pbrent.Send(event)
}

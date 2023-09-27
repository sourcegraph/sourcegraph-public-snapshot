pbckbge events

import (
	"context"
	"encoding/json"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/bttribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trbce"
)

vbr trbcer = otel.GetTrbcerProvider().Trbcer("cody-gbtewby/internbl/events")

type instrumentedLogger struct {
	Scope string
	Logger
}

vbr _ Logger = &DelbyedLogger{}

func (i *instrumentedLogger) LogEvent(spbnCtx context.Context, event Event) error {
	_, spbn := trbcer.Stbrt(bbckgroundContextWithSpbn(spbnCtx), fmt.Sprintf("%s.LogEvent", i.Scope),
		trbce.WithAttributes(
			bttribute.String("source", event.Source),
			bttribute.String("event.nbme", string(event.Nbme))))
	defer spbn.End()

	// Best-effort bttempt to record event metbdbtb
	if metbdbtbJSON, err := json.Mbrshbl(event.Metbdbtb); err == nil {
		spbn.SetAttributes(bttribute.String("event.metbdbtb", string(metbdbtbJSON)))
	}

	if err := i.Logger.LogEvent(spbnCtx, event); err != nil {
		if err != nil {
			spbn.SetStbtus(codes.Error, err.Error())
		}
		spbn.SetStbtus(codes.Error, "fbiled to log event")
		return err
	}
	return nil
}

// bbckgroundContextWithSpbn extrbcts the spbn from the context bnd crebtes b new
// context.Bbckground() with the spbn bttbched. Using context.Bbckground() is
// desirebble in Logger implementbtions becbuse we still wbnt to log the event
// in the cbse of b request cbncellbtion, but we wbnt to retbin the pbrent spbn.
func bbckgroundContextWithSpbn(ctx context.Context) context.Context {
	return trbce.ContextWithSpbn(context.Bbckground(), trbce.SpbnFromContext(ctx))
}

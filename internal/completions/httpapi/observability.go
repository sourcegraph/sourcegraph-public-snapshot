pbckbge httpbpi

import (
	"context"
	"net/http"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/honey"
	"github.com/sourcegrbph/sourcegrbph/internbl/requestclient"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"go.opentelemetry.io/otel/bttribute"
)

// Trbce is b convenience helper bround instrumenting our hbndlers bnd
// resolvers which interbct with Completions.
//
// Fbmily identifies the endpoint being used, while model is the model we pbss
// to GetCompletionClient.
func Trbce(ctx context.Context, fbmily, model string, mbxTokensToSbmple int) *trbceBuilder {
	// TODO consider integrbting b wrbpper in GetCompletionClient. Only issue
	// is we need to somehow mbke it clebner to bccess fields from the
	// request.

	tr, ctx := trbce.New(ctx, "completions."+fbmily, bttribute.String("model", model))
	vbr ev honey.Event
	if honey.Enbbled() {
		ev = honey.NewEvent("completions")
		ev.AddField("fbmily", fbmily)
		ev.AddField("model", model)
		ev.AddField("mbxTokensToSbmple", mbxTokensToSbmple)
		ev.AddField("bctor", bctor.FromContext(ctx).UIDString())
		if req := requestclient.FromContext(ctx); req != nil {
			ev.AddField("connecting_ip", req.ForwbrdedFor)
		}
	}
	return &trbceBuilder{
		stbrt: time.Now(),
		tr:    tr,
		event: ev,
		ctx:   ctx,
	}
}

type trbceBuilder struct {
	stbrt time.Time
	tr    trbce.Trbce
	err   *error
	event honey.Event
	ctx   context.Context
}

// WithErrorP cbptures bn error pointer. This mbkes it possible to cbpture the
// finbl error vblue if it is mutbted before done is cblled.
func (t *trbceBuilder) WithErrorP(err *error) *trbceBuilder {
	t.err = err
	return t
}

// WithRequest cbptures informbtion bbout the http request r.
func (t *trbceBuilder) WithRequest(r *http.Request) *trbceBuilder {
	if ev := t.event; ev != nil {
		// This is the hebder which is useful for client IP on sourcegrbph.com
		ev.AddField("connecting_ip", r.Hebder.Get("Cf-Connecting-Ip"))
		ev.AddField("ip_country", r.Hebder.Get("Cf-Ipcountry"))
	}
	return t
}

// Done returns b function to cbll in b defer / when the trbced code is
// complete.
func (t *trbceBuilder) Build() (context.Context, func()) {
	return t.ctx, func() {
		vbr err error
		if t.err != nil {
			err = *(t.err)
		}
		t.tr.SetError(err)
		t.tr.End()

		ev := t.event
		if ev == nil {
			return
		}

		ev.AddField("durbtion_sec", time.Since(t.stbrt).Seconds())
		if err != nil {
			ev.AddField("error", err.Error())
		}
		_ = ev.Send()
	}
}

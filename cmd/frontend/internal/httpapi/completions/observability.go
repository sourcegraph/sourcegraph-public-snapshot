package completions

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/requestclient"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// Trace is a convenience helper around instrumenting our handlers and
// resolvers which interact with Completions.
//
// Family identifies the endpoint being used, while model is the model we pass
// to GetCompletionClient.
func Trace(ctx context.Context, family, model string, maxTokensToSample int) *traceBuilder {
	// TODO consider integrating a wrapper in GetCompletionClient. Only issue
	// is we need to somehow make it cleaner to access fields from the
	// request.

	tr, ctx := trace.New(ctx, "completions."+family, attribute.String("model", model))
	var ev honey.Event
	if honey.Enabled() {
		ev = honey.NewEvent("completions")
		ev.AddField("family", family)
		ev.AddField("model", model)
		ev.AddField("maxTokensToSample", maxTokensToSample)
		ev.AddField("actor", actor.FromContext(ctx).UIDString())
		if req := requestclient.FromContext(ctx); req != nil {
			ev.AddField("connecting_ip", req.ForwardedFor)
		}
	}
	return &traceBuilder{
		start: time.Now(),
		tr:    tr,
		event: ev,
		ctx:   ctx,
	}
}

type traceBuilder struct {
	start time.Time
	tr    trace.Trace
	err   *error
	event honey.Event
	ctx   context.Context
}

// WithErrorP captures an error pointer. This makes it possible to capture the
// final error value if it is mutated before done is called.
func (t *traceBuilder) WithErrorP(err *error) *traceBuilder {
	t.err = err
	return t
}

// WithRequest captures information about the http request r.
func (t *traceBuilder) WithRequest(r *http.Request) *traceBuilder {
	if ev := t.event; ev != nil {
		// This is the header which is useful for client IP on sourcegraph.com
		ev.AddField("connecting_ip", r.Header.Get("Cf-Connecting-Ip"))
		ev.AddField("ip_country", r.Header.Get("Cf-Ipcountry"))
	}
	return t
}

// Done returns a function to call in a defer / when the traced code is
// complete.
func (t *traceBuilder) Build() (context.Context, func()) {
	return t.ctx, func() {
		var err error
		if t.err != nil {
			err = *(t.err)
		}
		t.tr.SetError(err)
		t.tr.End()

		ev := t.event
		if ev == nil {
			return
		}

		ev.AddField("duration_sec", time.Since(t.start).Seconds())
		if err != nil {
			ev.AddField("error", err.Error())
		}
		_ = ev.Send()
	}
}

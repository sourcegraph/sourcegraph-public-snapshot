package context

import (
	"src.sourcegraph.com/sourcegraph/vendored/github.com/resonancelabs/go-pub/instrument"

	"golang.org/x/net/context"
)

type contextKey int

const (
	activeSpanKey contextKey = iota
)

func activeSpan(ctx context.Context) instrument.ActiveSpan {
	val := ctx.Value(activeSpanKey)
	if span, ok := val.(instrument.ActiveSpan); ok {
		return span
	}
	return nil
}

func WithActiveSpan(ctx context.Context, span instrument.ActiveSpan) context.Context {
	return context.WithValue(ctx, activeSpanKey, span)
}

func StartSpan(parent context.Context) (context.Context, instrument.ActiveSpan) {
	newSpan := instrument.StartSpan()
	if oldSpan := activeSpan(parent); oldSpan != nil {
		for k, v := range oldSpan.TraceJoinIds() {
			newSpan.AddTraceJoinId(k, v)
		}
		newSpan.AddAttribute("parent_span_guid", string(oldSpan.Guid()))
	}
	ctx := WithActiveSpan(parent, newSpan)
	return ctx, newSpan
}

func FinishSpan(ctx context.Context) {
	if span := activeSpan(ctx); span != nil {
		span.Finish()
	} else {
		instrument.Log(instrument.Print("FinishSpan called without StartSpan").Payload(ctx))
	}
}

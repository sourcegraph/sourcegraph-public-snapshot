package xlang

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
)

func addTraceMeta(ctx context.Context) jsonrpc2.CallOption {
	carrier := opentracing.TextMapCarrier{}
	span := opentracing.SpanFromContext(ctx)
	if err := span.Tracer().Inject(span.Context(), opentracing.TextMap, carrier); err != nil {
		panic(err)
	}
	return jsonrpc2.Meta(carrier)
}

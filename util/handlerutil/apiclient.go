package handlerutil

import (
	"net/http"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
	"src.sourcegraph.com/sourcegraph/util/traceutil"
)

// TODO(sqs!): remove - just wraps sourcegraph.NewClientFromContext, which is a
// better design of this func.
var APIClient = func(r *http.Request) *sourcegraph.Client {
	// Add data to context that only exists on the request.
	ctx := httpctx.FromRequest(r)
	ctx = traceutil.NewContext(ctx, traceutil.SpanID(r))
	return sourcegraph.NewClientFromContext(ctx)
}

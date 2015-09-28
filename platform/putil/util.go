package putil

import (
	"net/http"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
)

func Context(r *http.Request) context.Context {
	return httpctx.FromRequest(r)
}

func UserFromRequest(r *http.Request) *sourcegraph.User {
	return handlerutil.UserFromRequest(r)
}

func UserFromContext(ctx context.Context) *sourcegraph.User {
	return handlerutil.UserFromContext(ctx)
}

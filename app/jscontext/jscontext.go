package jscontext

import (
	"net/http"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/app/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/buildvar"
	"sourcegraph.com/sourcegraph/sourcegraph/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/conf/feature"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/eventsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
	"sourcegraph.com/sourcegraph/sourcegraph/util/traceutil"
)

// JSContext is made available to JavaScript code via the
// "sourcegraph/app/context" module.
type JSContext struct {
	AppURL        string            `json:"appURL"`
	Authorization string            `json:"authorization"`
	CacheControl  string            `json:"cacheControl"`
	CurrentUser   *sourcegraph.User `json:"currentUser"`
	CurrentSpanID string            `json:"currentSpanID"`
	UserAgent     string            `json:"userAgent"`
	AssetsRoot    string            `json:"assetsRoot"`
	BuildVars     buildvar.Vars     `json:"buildVars"`
	Features      struct{}
}

// NewJSContextFromRequest populates a JSContext struct from the HTTP
// request.
func NewJSContextFromRequest(req *http.Request) (JSContext, error) {
	sess, err := auth.ReadSessionCookie(req)
	if err != nil && err != auth.ErrNoSession {
		return JSContext{}, err
	}

	// Propagate Cache-Control no-cache and max-age=0 directives
	// to the requests made by our client-side JavaScript. This is
	// not a perfect parser, but it catches the important cases.
	var cacheControl string
	if cc := req.Header.Get("cache-control"); strings.Contains(cc, "no-cache") || strings.Contains(cc, "max-age=0") {
		cacheControl = "no-cache"
	}

	ctx := JSContext{
		AppURL:        conf.AppURL(httpctx.FromRequest(req)).String(),
		CacheControl:  cacheControl,
		CurrentUser:   handlerutil.FullUserFromRequest(req),
		CurrentSpanID: traceutil.SpanID(req).String(),
		UserAgent:     eventsutil.UserAgentFromContext(httpctx.FromRequest(req)),
		AssetsRoot:    assets.URL("/").String(),
		BuildVars:     buildvar.All,
		Features:      feature.Features,
	}
	if sess != nil {
		ctx.Authorization = sess.AccessToken
	}

	return ctx, nil
}

package jscontext

import (
	"net/http"
	"strings"

	"github.com/justinas/nosurf"

	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/sourcegraph/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/app/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/buildvar"
	"sourcegraph.com/sourcegraph/sourcegraph/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/conf/feature"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
	"sourcegraph.com/sourcegraph/sourcegraph/util/traceutil"
)

type JSContext struct {
	AppURL        string
	Authorization string
	CacheControl  string
	CSRFToken     string
	CurrentUser   *sourcegraph.User
	CurrentSpanID appdash.SpanID
	DeviceID      string
	AssetsRoot    string
	BuildVars     buildvar.Vars
	Features      struct{}
}

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
		CSRFToken:     nosurf.Token(req),
		CurrentUser:   handlerutil.FullUserFromRequest(req),
		CurrentSpanID: traceutil.SpanID(req),
		DeviceID:      "", // TODO(pure-react)
		AssetsRoot:    assets.URL("/").String(),
		BuildVars:     buildvar.All,
		Features:      feature.Features,
	}
	if sess != nil {
		ctx.Authorization = sess.AccessToken
	}

	return ctx, nil
}

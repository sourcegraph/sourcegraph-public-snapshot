package jscontext

import (
	"net/http"
	"regexp"
	"strings"

	"context"

	"github.com/justinas/nosurf"

	"sourcegraph.com/sourcegraph/sourcegraph/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/app/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/buildvar"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/eventsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
)

// JSContext is made available to JavaScript code via the
// "sourcegraph/app/context" module.
type JSContext struct {
	AppURL         string        `json:"appURL"`
	AccessToken    string        `json:"accessToken"`
	CacheControl   string        `json:"cacheControl"`
	CSRFToken      string        `json:"csrfToken"`
	CurrentSpanID  string        `json:"currentSpanID"`
	UserAgentIsBot bool          `json:"userAgentIsBot"`
	AssetsRoot     string        `json:"assetsRoot"`
	BuildVars      buildvar.Vars `json:"buildVars"`
	Features       interface{}   `json:"features"`
}

// NewJSContextFromRequest populates a JSContext struct from the HTTP
// request.
func NewJSContextFromRequest(ctx context.Context, req *http.Request) (JSContext, error) {
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

	jsctx := JSContext{
		AppURL:         conf.AppURL(ctx).String(),
		CacheControl:   cacheControl,
		CSRFToken:      nosurf.Token(req),
		CurrentSpanID:  traceutil.SpanIDFromContext(ctx).String(),
		UserAgentIsBot: isBot(eventsutil.UserAgentFromContext(ctx)),
		AssetsRoot:     assets.URL("/").String(),
		BuildVars:      buildvar.Public,
		Features:       feature.Features,
	}
	if sess != nil {
		jsctx.AccessToken = sess.AccessToken
	}

	return jsctx, nil
}

var isBotPat = regexp.MustCompile(`(?i:googlecloudmonitoring|pingdom.com|go .* package http|sourcegraph e2etest|bot|crawl|slurp|spider|feed|rss|camo asset proxy|http-client|sourcegraph-client)`)

func isBot(userAgent string) bool {
	return isBotPat.MatchString(userAgent)
}

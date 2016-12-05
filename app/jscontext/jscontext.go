package jscontext

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/csrf"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/buildvar"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/eventsutil"
	httpapiauth "sourcegraph.com/sourcegraph/sourcegraph/services/httpapi/auth"
)

var sentryDSNFrontend = env.Get("SENTRY_DSN_FRONTEND", "", "Sentry/Raven DSN used for tracking of JavaScript errors")

// JSContext is made available to JavaScript code via the
// "sourcegraph/app/context" module.
type JSContext struct {
	AppURL            string                     `json:"appURL"`
	LegacyAccessToken string                     `json:"accessToken"` // Legacy support for Chrome extension.
	XHRHeaders        map[string]string          `json:"xhrHeaders"`
	CSRFToken         string                     `json:"csrfToken"`
	UserAgentIsBot    bool                       `json:"userAgentIsBot"`
	AssetsRoot        string                     `json:"assetsRoot"`
	BuildVars         buildvar.Vars              `json:"buildVars"`
	Features          interface{}                `json:"features"`
	User              *sourcegraph.User          `json:"user"`
	Emails            *sourcegraph.EmailAddrList `json:"emails"`
	GitHubToken       *sourcegraph.ExternalToken `json:"gitHubToken"`
	GoogleToken       *sourcegraph.ExternalToken `json:"googleToken"`
	SentryDSN         string                     `json:"sentryDSN"`
	IntercomHash      string                     `json:"intercomHash"`
}

// NewJSContextFromRequest populates a JSContext struct from the HTTP
// request.
func NewJSContextFromRequest(req *http.Request) (JSContext, error) {
	actor := auth.ActorFromContext(req.Context())

	headers := make(map[string]string)
	headers["x-sourcegraph-client"] = conf.AppURL.String()
	sessionCookie := auth.SessionCookie(req)
	if sessionCookie != "" {
		headers["Authorization"] = httpapiauth.AuthorizationHeaderWithSessionCookie(sessionCookie)
	}

	// -- currently we don't associate XHR calls with the parent page's span --
	// if span := opentracing.SpanFromContext(req.Context()); span != nil {
	// 	if err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.TextMapCarrier(headers)); err != nil {
	// 		return JSContext{}, err
	// 	}
	// }

	// Propagate Cache-Control no-cache and max-age=0 directives
	// to the requests made by our client-side JavaScript. This is
	// not a perfect parser, but it catches the important cases.
	if cc := req.Header.Get("cache-control"); strings.Contains(cc, "no-cache") || strings.Contains(cc, "max-age=0") {
		headers["Cache-Control"] = "no-cache"
	}

	csrfToken := csrf.Token(req)
	headers["X-Csrf-Token"] = csrfToken

	var gitHubToken *sourcegraph.ExternalToken
	if actor.GitHubConnected {
		gitHubToken = &sourcegraph.ExternalToken{
			Scope: strings.Join(actor.GitHubScopes, ","), // the UI only cares about the scope
		}
	}

	var googleToken *sourcegraph.ExternalToken
	if actor.GoogleConnected {
		googleToken = &sourcegraph.ExternalToken{
			Scope: strings.Join(actor.GoogleScopes, ","), // the UI only cares about the scope
		}
	}

	return JSContext{
		AppURL:            conf.AppURL.String(),
		LegacyAccessToken: sessionCookie, // Legacy support for Chrome extension.
		XHRHeaders:        headers,
		CSRFToken:         csrfToken,
		UserAgentIsBot:    isBot(eventsutil.UserAgentFromContext(req.Context())),
		AssetsRoot:        assets.URL("/").String(),
		BuildVars:         buildvar.Public,
		Features:          feature.Features,
		User:              actor.User(),
		Emails: &sourcegraph.EmailAddrList{
			EmailAddrs: []*sourcegraph.EmailAddr{{Email: actor.Email, Primary: true}},
		},
		GitHubToken:  gitHubToken,
		GoogleToken:  googleToken,
		SentryDSN:    sentryDSNFrontend,
		IntercomHash: intercomHMAC(actor.UID),
	}, nil
}

var isBotPat = regexp.MustCompile(`(?i:googlecloudmonitoring|pingdom.com|go .* package http|sourcegraph e2etest|bot|crawl|slurp|spider|feed|rss|camo asset proxy|http-client|sourcegraph-client)`)

func isBot(userAgent string) bool {
	return isBotPat.MatchString(userAgent)
}

var intercomSecretKey = env.Get("SG_INTERCOM_SECRET_KEY", "", "secret key for the Intercom widget")

func intercomHMAC(uid string) string {
	if uid == "" || intercomSecretKey == "" {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(intercomSecretKey))
	mac.Write([]byte(uid))
	return hex.EncodeToString(mac.Sum(nil))
}

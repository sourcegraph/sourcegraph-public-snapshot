package jscontext

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gorilla/csrf"
	opentracing "github.com/opentracing/opentracing-go"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/buildvar"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/eventsutil"
	httpapiauth "sourcegraph.com/sourcegraph/sourcegraph/services/httpapi/auth"
)

// JSContext is made available to JavaScript code via the
// "sourcegraph/app/context" module.
type JSContext struct {
	AppURL            string                     `json:"appURL"`
	LegacyAccessToken string                     `json:"accessToken"` // Legacy support for Chrome extension.
	XHRHeaders        map[string]string          `json:"xhrHeaders"`
	UserAgentIsBot    bool                       `json:"userAgentIsBot"`
	AssetsRoot        string                     `json:"assetsRoot"`
	BuildVars         buildvar.Vars              `json:"buildVars"`
	Features          interface{}                `json:"features"`
	User              *sourcegraph.User          `json:"user"`
	Emails            *sourcegraph.EmailAddrList `json:"emails"`
	GitHubToken       *sourcegraph.ExternalToken `json:"gitHubToken"`
	IntercomHash      string                     `json:"intercomHash"`
}

// NewJSContextFromRequest populates a JSContext struct from the HTTP
// request.
func NewJSContextFromRequest(req *http.Request) (JSContext, error) {
	actor := auth.ActorFromContext(req.Context())

	headers := make(map[string]string)
	sessionCookie := auth.SessionCookie(req)
	if sessionCookie != "" {
		headers["Authorization"] = httpapiauth.AuthorizationHeaderWithSessionCookie(sessionCookie)
	}

	if span := opentracing.SpanFromContext(req.Context()); span != nil {
		if err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.TextMapCarrier(headers)); err != nil {
			return JSContext{}, err
		}
	}

	// Propagate Cache-Control no-cache and max-age=0 directives
	// to the requests made by our client-side JavaScript. This is
	// not a perfect parser, but it catches the important cases.
	if cc := req.Header.Get("cache-control"); strings.Contains(cc, "no-cache") || strings.Contains(cc, "max-age=0") {
		headers["Cache-Control"] = "no-cache"
	}

	headers["X-Csrf-Token"] = csrf.Token(req)

	var gitHubToken *sourcegraph.ExternalToken
	if actor.GitHubConnected {
		gitHubToken = &sourcegraph.ExternalToken{
			Scope: strings.Join(actor.GitHubScopes, ","), // the UI only cares about the scope
		}
	}

	return JSContext{
		AppURL:            conf.AppURL.String(),
		LegacyAccessToken: sessionCookie, // Legacy support for Chrome extension.
		XHRHeaders:        headers,
		UserAgentIsBot:    isBot(eventsutil.UserAgentFromContext(req.Context())),
		AssetsRoot:        assets.URL("/").String(),
		BuildVars:         buildvar.Public,
		Features:          feature.Features,
		User:              actor.User(),
		Emails: &sourcegraph.EmailAddrList{
			EmailAddrs: []*sourcegraph.EmailAddr{{Email: actor.Email, Primary: true}},
		},
		GitHubToken:  gitHubToken,
		IntercomHash: intercomHMAC(actor.UID),
	}, nil
}

var isBotPat = regexp.MustCompile(`(?i:googlecloudmonitoring|pingdom.com|go .* package http|sourcegraph e2etest|bot|crawl|slurp|spider|feed|rss|camo asset proxy|http-client|sourcegraph-client)`)

func isBot(userAgent string) bool {
	return isBotPat.MatchString(userAgent)
}

var intercomSecretKey = os.Getenv("SG_INTERCOM_SECRET_KEY")

func intercomHMAC(uid string) string {
	if uid == "" || intercomSecretKey == "" {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(intercomSecretKey))
	mac.Write([]byte(uid))
	return hex.EncodeToString(mac.Sum(nil))
}

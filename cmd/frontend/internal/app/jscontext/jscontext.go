package jscontext

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/csrf"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/eventsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/stripe"
	httpapiauth "sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/auth"
)

var sentryDSNFrontend = env.Get("SENTRY_DSN_FRONTEND", "", "Sentry/Raven DSN used for tracking of JavaScript errors")
var authEnabledEnvVar = env.Get("AUTH_ENABLED", "true", "require login for users to view repositories")
var repoHomeRegexFilter = env.Get("REPO_HOME_REGEX_FILTER", "", "use this regex to filter for repositories on the repository landing page")

// TrackingAppID is used by the Telligent data pipeline
var TrackingAppID = env.Get("TRACKING_APP_ID", "", "application id to attribute front end user logs to. not providing this value will prevent logging.")

// JSContext is made available to JavaScript code via the
// "sourcegraph/app/context" module.
type JSContext struct {
	AppURL              string                     `json:"appURL"`
	LegacyAccessToken   string                     `json:"accessToken"` // Legacy support for Chrome extension.
	XHRHeaders          map[string]string          `json:"xhrHeaders"`
	CSRFToken           string                     `json:"csrfToken"`
	UserAgentIsBot      bool                       `json:"userAgentIsBot"`
	AssetsRoot          string                     `json:"assetsRoot"`
	Version             string                     `json:"version"`
	Features            interface{}                `json:"features"`
	User                *sourcegraph.User          `json:"user"`
	Emails              *sourcegraph.EmailAddrList `json:"emails"`
	GitHubToken         *sourcegraph.ExternalToken `json:"gitHubToken"`
	SentryDSN           string                     `json:"sentryDSN"`
	IntercomHash        string                     `json:"intercomHash"`
	TrackingAppID       string                     `json:"trackingAppID"`
	AuthEnabled         bool                       `json:"authEnabled"`
	RepoHomeRegexFilter string                     `json:"repoHomeRegexFilter"`
	StripePublicKey     string                     `json:"stripePublicKey"`
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

	var authEnabled, err = strconv.ParseBool(authEnabledEnvVar)
	if err != nil {
		log.Fatal(err)
		authEnabled = true
	}

	return JSContext{
		AppURL:            conf.AppURL.String(),
		LegacyAccessToken: sessionCookie, // Legacy support for Chrome extension.
		XHRHeaders:        headers,
		CSRFToken:         csrfToken,
		UserAgentIsBot:    isBot(eventsutil.UserAgentFromContext(req.Context())),
		AssetsRoot:        assets.URL("/").String(),
		Version:           env.Version,
		Features:          feature.Features,
		User:              actor.User(),
		Emails: &sourcegraph.EmailAddrList{
			EmailAddrs: []*sourcegraph.EmailAddr{{Email: actor.Email, Primary: true}},
		},
		GitHubToken:         gitHubToken,
		SentryDSN:           sentryDSNFrontend,
		IntercomHash:        intercomHMAC(actor.UID),
		AuthEnabled:         authEnabled,
		TrackingAppID:       TrackingAppID,
		RepoHomeRegexFilter: repoHomeRegexFilter,
		StripePublicKey:     stripe.StripePublicKey,
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

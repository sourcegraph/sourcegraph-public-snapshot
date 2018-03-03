package jscontext

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/csrf"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
	httpapiauth "sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/license"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/siteid"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

var sentryDSNFrontend = env.Get("SENTRY_DSN_FRONTEND", "", "Sentry/Raven DSN used for tracking of JavaScript errors")
var repoHomeRegexFilter = env.Get("REPO_HOME_REGEX_FILTER", "", "use this regex to filter for repositories on the repository landing page")

// immutableUser corresponds to the immutableUser type in the JS sourcegraphContext.
type immutableUser struct {
	UID        int32
	ExternalID *string `json:"externalID,omitempty"`
}

// JSContext is made available to JavaScript code via the
// "sourcegraph/app/context" module.
type JSContext struct {
	AppRoot        string            `json:"appRoot,omitempty"`
	AppURL         string            `json:"appURL,omitempty"`
	XHRHeaders     map[string]string `json:"xhrHeaders"`
	CSRFToken      string            `json:"csrfToken"`
	UserAgentIsBot bool              `json:"userAgentIsBot"`
	AssetsRoot     string            `json:"assetsRoot"`
	Version        string            `json:"version"`
	User           *immutableUser    `json:"user"`

	DisableTelemetry bool `json:"disableTelemetry"`

	GithubEnterpriseURLs map[string]string     `json:"githubEnterpriseURLs"`
	SentryDSN            string                `json:"sentryDSN"`
	SiteID               string                `json:"siteID"`
	Debug                bool                  `json:"debug"`
	RepoHomeRegexFilter  string                `json:"repoHomeRegexFilter"`
	SessionID            string                `json:"sessionID"`
	License              *license.License      `json:"license"`
	LicenseStatus        license.LicenseStatus `json:"licenseStatus"`
	ShowOnboarding       bool                  `json:"showOnboarding"`
	EmailEnabled         bool                  `json:"emailEnabled"`

	Site              schema.SiteConfiguration `json:"site"` // public subset of site configuration
	LikelyDockerOnMac bool                     `json:"likelyDockerOnMac"`

	SourcegraphDotComMode bool `json:"sourcegraphDotComMode"`
}

// NewJSContextFromRequest populates a JSContext struct from the HTTP
// request.
func NewJSContextFromRequest(req *http.Request) JSContext {
	actor := actor.FromContext(req.Context())

	headers := make(map[string]string)
	headers["x-sourcegraph-client"] = globals.AppURL.String()
	sessionCookie := session.SessionCookie(req)
	sessionID := httpapiauth.AuthorizationHeaderWithSessionCookie(sessionCookie)
	if sessionCookie != "" {
		headers["Authorization"] = sessionID
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

	var user *immutableUser
	if actor.IsAuthenticated() {
		user = &immutableUser{UID: actor.UID}

		if u, err := db.Users.GetByID(req.Context(), actor.UID); err == nil && u != nil {
			user.ExternalID = u.ExternalID
		}
	}

	siteID := siteid.Get()

	// For legacy configurations that have a license key already set we should not overwrite their existing configuration details.
	license, licenseStatus := license.Get(siteID)
	var showOnboarding = false
	if license == nil || license.SiteID == "" {
		siteConfig, err := db.SiteConfig.Get(req.Context())
		showOnboarding = err == nil && !siteConfig.Initialized
	}

	return JSContext{
		AppURL:               globals.AppURL.String(),
		XHRHeaders:           headers,
		CSRFToken:            csrfToken,
		UserAgentIsBot:       isBot(req.UserAgent()),
		AssetsRoot:           assets.URL("").String(),
		Version:              env.Version,
		User:                 user,
		DisableTelemetry:     conf.GetTODO().DisableTelemetry,
		GithubEnterpriseURLs: conf.GitHubEnterpriseURLs(),
		SentryDSN:            sentryDSNFrontend,
		Debug:                envvar.DebugMode(),
		SiteID:               siteID,
		RepoHomeRegexFilter:  repoHomeRegexFilter,
		SessionID:            sessionID,
		License:              license,
		LicenseStatus:        licenseStatus,
		ShowOnboarding:       showOnboarding,
		EmailEnabled:         conf.CanSendEmail(),
		Site:                 publicSiteConfiguration,
		LikelyDockerOnMac:    likelyDockerOnMac(),

		SourcegraphDotComMode: envvar.SourcegraphDotComMode(),
	}
}

// publicSiteConfiguration is the subset of the site.schema.json site configuration
// that is necessary for the web app and is not sensitive/secret.
var publicSiteConfiguration = schema.SiteConfiguration{
	AuthAllowSignup: conf.GetTODO().AuthAllowSignup,
}

var isBotPat = regexp.MustCompile(`(?i:googlecloudmonitoring|pingdom.com|go .* package http|sourcegraph e2etest|bot|crawl|slurp|spider|feed|rss|camo asset proxy|http-client|sourcegraph-client)`)

func isBot(userAgent string) bool {
	return isBotPat.MatchString(userAgent)
}

func likelyDockerOnMac() bool {
	data, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		return false // permission errors, or maybe not a Linux OS, etc. Assume we're not docker for mac.
	}
	return bytes.Contains(data, []byte("mac")) || bytes.Contains(data, []byte("osx"))
}

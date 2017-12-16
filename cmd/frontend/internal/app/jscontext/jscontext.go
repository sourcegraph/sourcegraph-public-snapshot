package jscontext

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/notif"

	"github.com/gorilla/csrf"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth0"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
	httpapiauth "sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/license"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

var sentryDSNFrontend = env.Get("SENTRY_DSN_FRONTEND", "", "Sentry/Raven DSN used for tracking of JavaScript errors")
var repoHomeRegexFilter = env.Get("REPO_HOME_REGEX_FILTER", "", "use this regex to filter for repositories on the repository landing page")

// TrackingAppID is used by the Telligent data pipeline
var TrackingAppID = conf.Get().AppID

var gitHubAppURL = env.Get("SRC_GITHUB_APP_URL", "", "URL for the GitHub app landing page users are taken to after being prompted to install the Sourcegraph GitHub app.")
var githubConf = conf.Get().GitHub

// githubEnterpriseURLs is a map of GitHub Enerprise hosts to their full URLs.
// This can be used for the purposes of generating external GitHub enterprise links.
var githubEnterpriseURLs = make(map[string]string)

func init() {
	for _, c := range githubConf {
		gheURL, err := url.Parse(c.URL)
		if err != nil {
			log15.Error("error parsing GitHub config", "error", err)
		}
		githubEnterpriseURLs[gheURL.Host] = strings.TrimSuffix(c.URL, "/")
	}
}

// immutableUser corresponds to the immutableUser type in the JS sourcegraphContext.
type immutableUser struct {
	UID     string
	IsAdmin bool
}

// JSContext is made available to JavaScript code via the
// "sourcegraph/app/context" module.
type JSContext struct {
	AppRoot        string            `json:"appRoot,omitempty"`
	AppURL         string            `json:"appURL,omitempty"`
	AppID          string            `json:"appID,omitempty"`
	XHRHeaders     map[string]string `json:"xhrHeaders"`
	CSRFToken      string            `json:"csrfToken"`
	UserAgentIsBot bool              `json:"userAgentIsBot"`
	AssetsRoot     string            `json:"assetsRoot"`
	Version        string            `json:"version"`
	Features       interface{}       `json:"features"`
	User           *immutableUser    `json:"user"`

	GitHubToken          *sourcegraph.ExternalToken `json:"gitHubToken"`
	GitHubAppURL         string                     `json:"gitHubAppURL"`
	GithubEnterpriseURLs map[string]string          `json:"githubEnterpriseURLs"`
	SentryDSN            string                     `json:"sentryDSN"`
	IntercomHash         string                     `json:"intercomHash"`
	TrackingAppID        string                     `json:"trackingAppID"`
	Debug                bool                       `json:"debug"`
	OnPrem               bool                       `json:"onPrem"`
	RepoHomeRegexFilter  string                     `json:"repoHomeRegexFilter"`
	SessionID            string                     `json:"sessionID"`
	Auth0Domain          string                     `json:"auth0Domain"`
	Auth0ClientID        string                     `json:"auth0ClientID"`
	License              *license.License           `json:"license"`
	LicenseStatus        license.LicenseStatus      `json:"licenseStatus"`
	ShowOnboarding       bool                       `json:"showOnboarding"`
	EmailEnabled         bool                       `json:"emailEnabled"`
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

	var gitHubToken *sourcegraph.ExternalToken
	if actor.GitHubConnected {
		gitHubToken = &sourcegraph.ExternalToken{
			Scope: strings.Join(actor.GitHubScopes, ","), // the UI only cares about the scope
		}
	}

	var user *immutableUser
	if actor.UID != "" && actor != nil {
		user = &immutableUser{UID: actor.UID, IsAdmin: actor.IsAdmin()}
	}

	// For legacy configurations that have a license key already set we should not overwrite their existing configuration details.
	license, licenseStatus := license.Get(TrackingAppID)
	var showOnboarding = false
	if envvar.DeploymentOnPrem() && (license == nil || license.AppID == "") {
		deploymentConfiguration, err := store.DeploymentConfiguration.Get(req.Context())
		if err != nil {
			// errors swallowed because telemetry is optional.
			log15.Error("store.Config.Get failed", "error", err)
		} else if deploymentConfiguration.TelemetryEnabled {
			TrackingAppID = deploymentConfiguration.AppID
		} else {
			TrackingAppID = ""
		}
		showOnboarding = deploymentConfiguration.LastUpdated == ""
	}

	return JSContext{
		AppURL:               globals.AppURL.String(),
		XHRHeaders:           headers,
		CSRFToken:            csrfToken,
		UserAgentIsBot:       isBot(req.UserAgent()),
		AssetsRoot:           assets.URL("/").String(),
		Version:              env.Version,
		Features:             feature.Features,
		User:                 user,
		GitHubToken:          gitHubToken,
		GitHubAppURL:         gitHubAppURL,
		GithubEnterpriseURLs: githubEnterpriseURLs,
		SentryDSN:            sentryDSNFrontend,
		IntercomHash:         intercomHMAC(actor.UID),
		Debug:                envvar.DebugMode(),
		OnPrem:               envvar.DeploymentOnPrem(),
		TrackingAppID:        TrackingAppID,
		RepoHomeRegexFilter:  repoHomeRegexFilter,
		SessionID:            sessionID,
		Auth0Domain:          auth0.Domain,
		Auth0ClientID:        auth0.Config.ClientID,
		License:              license,
		LicenseStatus:        licenseStatus,
		ShowOnboarding:       showOnboarding,
		EmailEnabled:         notif.EmailIsConfigured(),
	}
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

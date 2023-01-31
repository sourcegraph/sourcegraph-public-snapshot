// Package jscontext contains functionality for information we pass down into
// the JS webapp.
package jscontext

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/hooks"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assetsutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/userpasswd"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/siteid"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/schema"
)

// BillingPublishableKey is the publishable (non-secret) API key for the billing system, if any.
var BillingPublishableKey string

type authProviderInfo struct {
	IsBuiltin         bool   `json:"isBuiltin"`
	DisplayName       string `json:"displayName"`
	ServiceType       string `json:"serviceType"`
	AuthenticationURL string `json:"authenticationURL"`
	ServiceID         string `json:"serviceID"`
}

// GenericPasswordPolicy a generic password policy that holds password requirements
type authPasswordPolicy struct {
	Enabled                   bool `json:"enabled"`
	NumberOfSpecialCharacters int  `json:"numberOfSpecialCharacters"`
	RequireAtLeastOneNumber   bool `json:"requireAtLeastOneNumber"`
	RequireUpperAndLowerCase  bool `json:"requireUpperAndLowerCase"`
}
type UserLatestSettings struct {
	ID       int32  // the unique ID of this settings value
	Contents string // the raw JSON (with comments and trailing commas allowed)
}
type UserOrganization struct {
	ID          graphql.ID
	Name        string
	DisplayName *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CurrentUser struct {
	ID                  graphql.ID
	DatabaseID          int32
	Username            string
	AvatarURL           string
	DisplayName         string
	SiteAdmin           bool
	URL                 string
	SettingsURL         string
	ViewerCanAdminister bool
	Tags                []string
	TosAccepted         bool
	Searchable          bool

	Organizations  []*UserOrganization
	CanSignOut     *bool
	Emails         []*database.UserEmail
	LatestSettings *UserLatestSettings
}

// JSContext is made available to JavaScript code via the
// "sourcegraph/app/context" module.
//
// 🚨 SECURITY: This struct is sent to all users regardless of whether or
// not they are logged in, for example on an auth.public=false private
// server. Including secret fields here is OK if it is based on the user's
// authentication above, but do not include e.g. hard-coded secrets about
// the server instance here as they would be sent to anonymous users.
type JSContext struct {
	AppRoot        string            `json:"appRoot,omitempty"`
	ExternalURL    string            `json:"externalURL,omitempty"`
	XHRHeaders     map[string]string `json:"xhrHeaders"`
	UserAgentIsBot bool              `json:"userAgentIsBot"`
	AssetsRoot     string            `json:"assetsRoot"`
	Version        string            `json:"version"`

	IsAuthenticatedUser bool         `json:"isAuthenticatedUser"`
	CurrentUser         *CurrentUser `json:"CurrentUser"`

	SentryDSN     *string               `json:"sentryDSN"`
	OpenTelemetry *schema.OpenTelemetry `json:"openTelemetry"`

	SiteID        string `json:"siteID"`
	SiteGQLID     string `json:"siteGQLID"`
	Debug         bool   `json:"debug"`
	NeedsSiteInit bool   `json:"needsSiteInit"`
	EmailEnabled  bool   `json:"emailEnabled"`

	Site              schema.SiteConfiguration `json:"site"` // public subset of site configuration
	LikelyDockerOnMac bool                     `json:"likelyDockerOnMac"`
	NeedServerRestart bool                     `json:"needServerRestart"`
	DeployType        string                   `json:"deployType"`

	SourcegraphDotComMode bool `json:"sourcegraphDotComMode"`

	BillingPublishableKey string `json:"billingPublishableKey,omitempty"`

	AccessTokensAllow conf.AccessTokenAllow `json:"accessTokensAllow"`

	AllowSignup bool `json:"allowSignup"`

	ResetPasswordEnabled bool `json:"resetPasswordEnabled"`

	ExternalServicesUserMode string `json:"externalServicesUserMode"`

	AuthMinPasswordLength int                `json:"authMinPasswordLength"`
	AuthPasswordPolicy    authPasswordPolicy `json:"authPasswordPolicy"`

	AuthProviders []authProviderInfo `json:"authProviders"`

	Branding *schema.Branding `json:"branding"`

	BatchChangesEnabled                bool `json:"batchChangesEnabled"`
	BatchChangesDisableWebhooksWarning bool `json:"batchChangesDisableWebhooksWarning"`
	BatchChangesWebhookLogsEnabled     bool `json:"batchChangesWebhookLogsEnabled"`

	ExecutorsEnabled                         bool `json:"executorsEnabled"`
	CodeIntelAutoIndexingEnabled             bool `json:"codeIntelAutoIndexingEnabled"`
	CodeIntelAutoIndexingAllowGlobalPolicies bool `json:"codeIntelAutoIndexingAllowGlobalPolicies"`

	CodeInsightsEnabled bool `json:"codeInsightsEnabled"`

	RedirectUnsupportedBrowser bool `json:"RedirectUnsupportedBrowser"`

	ProductResearchPageEnabled bool `json:"productResearchPageEnabled"`

	ExperimentalFeatures schema.ExperimentalFeatures `json:"experimentalFeatures"`

	EnableLegacyExtensions bool `json:"enableLegacyExtensions"`

	LicenseInfo *hooks.LicenseInfo `json:"licenseInfo"`

	OutboundRequestLogLimit int `json:"outboundRequestLogLimit"`

	DisableFeedbackSurvey bool `json:"disableFeedbackSurvey"`
}

// NewJSContextFromRequest populates a JSContext struct from the HTTP
// request.
func NewJSContextFromRequest(req *http.Request, db database.DB) JSContext {
	ctx := req.Context()
	a := actor.FromContext(ctx)

	headers := make(map[string]string)
	headers["x-sourcegraph-client"] = globals.ExternalURL().String()
	headers["X-Requested-With"] = "Sourcegraph" // required for httpapi to use cookie auth

	// Propagate Cache-Control no-cache and max-age=0 directives
	// to the requests made by our client-side JavaScript. This is
	// not a perfect parser, but it catches the important cases.
	if cc := req.Header.Get("cache-control"); strings.Contains(cc, "no-cache") || strings.Contains(cc, "max-age=0") {
		headers["Cache-Control"] = "no-cache"
	}

	siteID := siteid.Get()

	// Show the site init screen?
	globalState, err := db.GlobalState().Get(ctx)
	needsSiteInit := err == nil && !globalState.Initialized

	// Auth providers
	var authProviders []authProviderInfo
	for _, p := range providers.Providers() {
		if p.Config().Github != nil && p.Config().Github.Hidden {
			continue
		}
		info := p.CachedInfo()
		if info != nil {
			authProviders = append(authProviders, authProviderInfo{
				IsBuiltin:         p.Config().Builtin != nil,
				DisplayName:       info.DisplayName,
				ServiceType:       p.ConfigID().Type,
				AuthenticationURL: info.AuthenticationURL,
				ServiceID:         info.ServiceID,
			})
		}
	}

	pp := conf.AuthPasswordPolicy()

	var authPasswordPolicy authPasswordPolicy
	authPasswordPolicy.Enabled = pp.Enabled
	authPasswordPolicy.NumberOfSpecialCharacters = pp.NumberOfSpecialCharacters
	authPasswordPolicy.RequireAtLeastOneNumber = pp.RequireAtLeastOneNumber
	authPasswordPolicy.RequireUpperAndLowerCase = pp.RequireUpperandLowerCase

	var sentryDSN *string
	siteConfig := conf.Get().SiteConfiguration

	if siteConfig.Log != nil && siteConfig.Log.Sentry != nil && siteConfig.Log.Sentry.Dsn != "" {
		sentryDSN = &siteConfig.Log.Sentry.Dsn
	}

	var openTelemetry *schema.OpenTelemetry
	if clientObservability := siteConfig.ObservabilityClient; clientObservability != nil {
		openTelemetry = clientObservability.OpenTelemetry
	}

	var licenseInfo *hooks.LicenseInfo
	var user *types.User
	if !a.IsAuthenticated() {
		licenseInfo = hooks.GetLicenseInfo(false)
	} else {
		// Ignore err as we don't care if user does not exist
		user, _ = a.User(ctx, db.Users())
		licenseInfo = hooks.GetLicenseInfo(user != nil && user.SiteAdmin)
	}

	// 🚨 SECURITY: This struct is sent to all users regardless of whether or
	// not they are logged in, for example on an auth.public=false private
	// server. Including secret fields here is OK if it is based on the user's
	// authentication above, but do not include e.g. hard-coded secrets about
	// the server instance here as they would be sent to anonymous users.
	return JSContext{
		ExternalURL:         globals.ExternalURL().String(),
		XHRHeaders:          headers,
		UserAgentIsBot:      isBot(req.UserAgent()),
		AssetsRoot:          assetsutil.URL("").String(),
		Version:             version.Version(),
		IsAuthenticatedUser: a.IsAuthenticated(),
		CurrentUser:         createCurrentUser(ctx, user, db),

		SentryDSN:                  sentryDSN,
		OpenTelemetry:              openTelemetry,
		RedirectUnsupportedBrowser: siteConfig.RedirectUnsupportedBrowser,
		Debug:                      env.InsecureDev,
		SiteID:                     siteID,

		SiteGQLID: string(graphqlbackend.SiteGQLID()),

		NeedsSiteInit:     needsSiteInit,
		EmailEnabled:      conf.CanSendEmail(),
		Site:              publicSiteConfiguration(),
		LikelyDockerOnMac: likelyDockerOnMac(),
		NeedServerRestart: globals.ConfigurationServerFrontendOnly.NeedServerRestart(),
		DeployType:        deploy.Type(),

		SourcegraphDotComMode: envvar.SourcegraphDotComMode(),

		BillingPublishableKey: BillingPublishableKey,

		// Experiments. We pass these through explicitly, so we can
		// do the default behavior only in Go land.
		AccessTokensAllow: conf.AccessTokensAllow(),

		ResetPasswordEnabled: userpasswd.ResetPasswordEnabled(),

		ExternalServicesUserMode: conf.ExternalServiceUserMode().String(),

		AllowSignup: conf.AuthAllowSignup(),

		AuthMinPasswordLength: conf.AuthMinPasswordLength(),
		AuthPasswordPolicy:    authPasswordPolicy,

		AuthProviders: authProviders,

		Branding: globals.Branding(),

		BatchChangesEnabled:                enterprise.BatchChangesEnabledForUser(ctx, db) == nil,
		BatchChangesDisableWebhooksWarning: conf.Get().BatchChangesDisableWebhooksWarning,
		BatchChangesWebhookLogsEnabled:     webhooks.LoggingEnabled(conf.Get()),

		ExecutorsEnabled:                         conf.ExecutorsEnabled(),
		CodeIntelAutoIndexingEnabled:             conf.CodeIntelAutoIndexingEnabled(),
		CodeIntelAutoIndexingAllowGlobalPolicies: conf.CodeIntelAutoIndexingAllowGlobalPolicies(),

		CodeInsightsEnabled: enterprise.IsCodeInsightsEnabled(),

		ProductResearchPageEnabled: conf.ProductResearchPageEnabled(),

		ExperimentalFeatures: conf.ExperimentalFeatures(),

		EnableLegacyExtensions: conf.ExperimentalFeatures().EnableLegacyExtensions,

		LicenseInfo: licenseInfo,

		OutboundRequestLogLimit: conf.Get().OutboundRequestLogLimit,

		DisableFeedbackSurvey: conf.Get().DisableFeedbackSurvey,
	}
}

// createCurrentUser creates CurrentUser object which contains of types.User
// properties along with some extra data such as user emails, organisations,
// session information, etc.
func createCurrentUser(ctx context.Context, user *types.User, db database.DB) *CurrentUser {
	if user == nil {
		return nil
	}
	url := fmt.Sprintf("/users/%s", user.Username)
	settingsURL := fmt.Sprintf("%s/settings", url)

	return &CurrentUser{
		ID: relay.MarshalID("User", user.ID),
		// DatabaseID is just a user ID
		DatabaseID:          user.ID,
		Username:            user.Username,
		AvatarURL:           user.AvatarURL,
		DisplayName:         user.DisplayName,
		SiteAdmin:           user.SiteAdmin,
		URL:                 url,
		SettingsURL:         settingsURL,
		ViewerCanAdminister: resolveViewerCanAdminister(ctx, user, db),
		Tags:                user.Tags,
		TosAccepted:         user.TosAccepted,
		Searchable:          user.Searchable,
		Organizations:       resolveUserOrgs(ctx, user, db),
		CanSignOut:          resolveUserCanSignOut(ctx, user),
		Emails:              resolveUserEmails(ctx, user, db),
		LatestSettings:      resolveLatestSettings(ctx, user, db),
	}
}

func resolveViewerCanAdminister(ctx context.Context, user *types.User, db database.DB) bool {
	// 🚨 SECURITY: Only the authenticated user can administrate themselves on
	// Sourcegraph.com.
	var err error
	if envvar.SourcegraphDotComMode() {
		err = auth.CheckSameUser(ctx, user.ID)
	} else {
		err = auth.CheckSiteAdminOrSameUser(ctx, db, user.ID)
	}
	if envvar.SourcegraphDotComMode() {
		if err := auth.CheckSameUser(ctx, user.ID); err != nil {
			return false
		}
	} else {
		if err := auth.CheckSiteAdminOrSameUser(ctx, db, user.ID); err != nil {
			return false
		}
	}
	if errcode.IsUnauthorized(err) {
		return false
	} else if err != nil {
		return false
	}
	return true
}

func resolveUserOrgs(ctx context.Context, user *types.User, db database.DB) []*UserOrganization {
	// 🚨 SECURITY: Only the user and admins are allowed to access user
	// organisations.
	if err := auth.CheckSiteAdminOrSameUser(ctx, db, user.ID); err != nil {
		return nil
	}
	orgs, err := db.Orgs().GetByUserID(ctx, user.ID)
	if err != nil {
		return nil
	}
	userOrganizations := make([]*UserOrganization, 0, len(orgs))
	for _, org := range orgs {
		userOrganizations = append(userOrganizations, convertOrgToUserOrganization(org))
	}
	return userOrganizations
}

func convertOrgToUserOrganization(org *types.Org) *UserOrganization {
	return &UserOrganization{
		ID:          relay.MarshalID("Org", org.ID),
		Name:        org.Name,
		DisplayName: org.DisplayName,
		CreatedAt:   org.CreatedAt,
		UpdatedAt:   org.UpdatedAt,
	}
}

func resolveUserCanSignOut(ctx context.Context, user *types.User) *bool {
	// 🚨 SECURITY: Only the user can view their session information, because it is
	// retrieved from the context of this request (and not persisted in a way that is
	// queryable).
	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() || a.UID != user.ID {
		return nil
	}

	canSignOut := false
	if a.FromSessionCookie {
		// The http-header auth provider is the only auth provider that a user cannot
		// sign out from.
		for _, p := range conf.Get().AuthProviders {
			if p.HttpHeader == nil {
				canSignOut = true
				break
			}
		}
	}
	return &canSignOut
}

func resolveUserEmails(ctx context.Context, user *types.User, db database.DB) []*database.UserEmail {
	// 🚨 SECURITY: Only the authenticated user and site admins can list user's
	// emails.
	if err := auth.CheckSiteAdminOrSameUser(ctx, db, user.ID); err != nil {
		return nil
	}

	userEmails, err := db.UserEmails().ListByUser(ctx, database.UserEmailsListOptions{
		UserID: user.ID,
	})
	if err != nil {
		return nil
	}
	return userEmails
}

func resolveLatestSettings(ctx context.Context, user *types.User, db database.DB) *UserLatestSettings {
	// 🚨 SECURITY: Only the authenticated user can view their settings on
	// Sourcegraph.com.
	if envvar.SourcegraphDotComMode() {
		if err := auth.CheckSameUser(ctx, user.ID); err != nil {
			return nil
		}
	} else {
		// 🚨 SECURITY: Only the user and admins are allowed to access other user
		// settings, because they may contain secrets or other sensitive data.
		if err := auth.CheckSiteAdminOrSameUser(ctx, db, user.ID); err != nil {
			return nil
		}
	}

	settings, err := db.Settings().GetLatest(ctx, api.SettingsSubject{User: &user.ID})
	if err != nil {
		return nil
	}
	if settings == nil {
		return nil
	}

	return &UserLatestSettings{ID: settings.ID, Contents: settings.Contents}
}

// publicSiteConfiguration is the subset of the site.schema.json site
// configuration that is necessary for the web app and is not sensitive/secret.
func publicSiteConfiguration() schema.SiteConfiguration {
	c := conf.Get()
	updateChannel := c.UpdateChannel
	if updateChannel == "" {
		updateChannel = "release"
	}
	return schema.SiteConfiguration{
		AuthPublic:                  c.AuthPublic,
		UpdateChannel:               updateChannel,
		AuthzEnforceForSiteAdmins:   c.AuthzEnforceForSiteAdmins,
		DisableNonCriticalTelemetry: c.DisableNonCriticalTelemetry,
	}
}

var isBotPat = lazyregexp.New(`(?i:googlecloudmonitoring|pingdom.com|go .* package http|sourcegraph e2etest|bot|crawl|slurp|spider|feed|rss|camo asset proxy|http-client|sourcegraph-client)`)

func isBot(userAgent string) bool {
	return isBotPat.MatchString(userAgent)
}

func likelyDockerOnMac() bool {
	r := net.DefaultResolver
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	addrs, err := r.LookupHost(ctx, "host.docker.internal")
	if err != nil || len(addrs) == 0 {
		return false //  Assume we're not docker for mac.
	}
	return true
}

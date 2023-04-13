// Package jscontext contains functionality for information we pass down into
// the JS webapp.
package jscontext

import (
	"context"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	logger "github.com/sourcegraph/log"

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
	sgactor "github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/singleprogram/filepicker"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/schema"
)

// BillingPublishableKey is the publishable (non-secret) API key for the billing system, if any.
var BillingPublishableKey string

type authProviderInfo struct {
	IsBuiltin         bool    `json:"isBuiltin"`
	DisplayName       string  `json:"displayName"`
	DisplayPrefix     *string `json:"displayPrefix"`
	ServiceType       string  `json:"serviceType"`
	AuthenticationURL string  `json:"authenticationURL"`
	ServiceID         string  `json:"serviceID"`
}

// GenericPasswordPolicy a generic password policy that holds password requirements
type authPasswordPolicy struct {
	Enabled                   bool `json:"enabled"`
	NumberOfSpecialCharacters int  `json:"numberOfSpecialCharacters"`
	RequireAtLeastOneNumber   bool `json:"requireAtLeastOneNumber"`
	RequireUpperAndLowerCase  bool `json:"requireUpperAndLowerCase"`
}
type UserLatestSettings struct {
	ID       int32                      `json:"id"`       // the unique ID of this settings value
	Contents graphqlbackend.JSONCString `json:"contents"` // the raw JSON (with comments and trailing commas allowed)
}
type UserOrganization struct {
	Typename    string     `json:"__typename"`
	ID          graphql.ID `json:"id"`
	Name        string     `json:"name"`
	DisplayName *string    `json:"displayName"`
	URL         string     `json:"url"`
	SettingsURL *string    `json:"settingsURL"`
}
type UserOrganizationsConnection struct {
	Typename string             `json:"__typename"`
	Nodes    []UserOrganization `json:"nodes"`
}
type UserEmail struct {
	Email     string `json:"email"`
	IsPrimary bool   `json:"isPrimary"`
	Verified  bool   `json:"verified"`
}
type UserSession struct {
	CanSignOut bool `json:"canSignOut"`
}

type TemporarySettings struct {
	GraphQLTypename string `json:"__typename"`
	Contents        string `json:"contents"`
}

type Permission struct {
	GraphQLTypename string     `json:"__typename"`
	ID              graphql.ID `json:"id"`
	DisplayName     string     `json:"displayName"`
}

type PermissionsConnection struct {
	GraphQLTypename string       `json:"__typename"`
	Nodes           []Permission `json:"nodes"`
}
type CurrentUser struct {
	GraphQLTypename     string     `json:"__typename"`
	ID                  graphql.ID `json:"id"`
	DatabaseID          int32      `json:"databaseID"`
	Username            string     `json:"username"`
	AvatarURL           *string    `json:"avatarURL"`
	DisplayName         string     `json:"displayName"`
	SiteAdmin           bool       `json:"siteAdmin"`
	URL                 string     `json:"url"`
	SettingsURL         string     `json:"settingsURL"`
	ViewerCanAdminister bool       `json:"viewerCanAdminister"`
	TosAccepted         bool       `json:"tosAccepted"`
	Searchable          bool       `json:"searchable"`

	Organizations  *UserOrganizationsConnection `json:"organizations"`
	Session        *UserSession                 `json:"session"`
	Emails         []UserEmail                  `json:"emails"`
	LatestSettings *UserLatestSettings          `json:"latestSettings"`
	Permissions    PermissionsConnection        `json:"permissions"`
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

	IsAuthenticatedUser bool               `json:"isAuthenticatedUser"`
	CurrentUser         *CurrentUser       `json:"currentUser"`
	TemporarySettings   *TemporarySettings `json:"temporarySettings"`

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
	SourcegraphAppMode    bool `json:"sourcegraphAppMode"`

	BillingPublishableKey string `json:"billingPublishableKey,omitempty"`

	AccessTokensAllow conf.AccessTokenAllow `json:"accessTokensAllow"`

	AllowSignup bool `json:"allowSignup"`

	ResetPasswordEnabled bool `json:"resetPasswordEnabled"`

	ExternalServicesUserMode string `json:"externalServicesUserMode"`

	AuthMinPasswordLength int                `json:"authMinPasswordLength"`
	AuthPasswordPolicy    authPasswordPolicy `json:"authPasswordPolicy"`

	AuthProviders                  []authProviderInfo `json:"authProviders"`
	AuthPrimaryLoginProvidersCount int                `json:"primaryLoginProvidersCount"`

	AuthAccessRequest *schema.AuthAccessRequest `json:"authAccessRequest"`

	Branding *schema.Branding `json:"branding"`

	BatchChangesEnabled                bool                                `json:"batchChangesEnabled"`
	BatchChangesDisableWebhooksWarning bool                                `json:"batchChangesDisableWebhooksWarning"`
	BatchChangesWebhookLogsEnabled     bool                                `json:"batchChangesWebhookLogsEnabled"`
	BatchChangesRolloutWindows         *[]*schema.BatchChangeRolloutWindow `json:"batchChangesRolloutWindows"`

	ExecutorsEnabled                         bool `json:"executorsEnabled"`
	CodeIntelAutoIndexingEnabled             bool `json:"codeIntelAutoIndexingEnabled"`
	CodeIntelAutoIndexingAllowGlobalPolicies bool `json:"codeIntelAutoIndexingAllowGlobalPolicies"`

	CodeInsightsEnabled bool `json:"codeInsightsEnabled"`

	EmbeddingsEnabled bool `json:"embeddingsEnabled"`

	RedirectUnsupportedBrowser bool `json:"RedirectUnsupportedBrowser"`

	ProductResearchPageEnabled bool `json:"productResearchPageEnabled"`

	ExperimentalFeatures schema.ExperimentalFeatures `json:"experimentalFeatures"`

	LicenseInfo *hooks.LicenseInfo `json:"licenseInfo"`

	OutboundRequestLogLimit int `json:"outboundRequestLogLimit"`

	DisableFeedbackSurvey bool `json:"disableFeedbackSurvey"`

	NeedsRepositoryConfiguration bool `json:"needsRepositoryConfiguration"`

	ExtsvcConfigFileExists bool `json:"extsvcConfigFileExists"`

	ExtsvcConfigAllowEdits bool `json:"extsvcConfigAllowEdits"`

	RunningOnMacOS bool `json:"runningOnMacOS"`

	LocalFilePickerAvailable bool `json:"localFilePickerAvailable"`

	SrcServeGitUrl string `json:"srcServeGitUrl"`
}

// NewJSContextFromRequest populates a JSContext struct from the HTTP
// request.
func NewJSContextFromRequest(req *http.Request, db database.DB) JSContext {
	ctx := req.Context()
	a := sgactor.FromContext(ctx)

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
	siteInitialized, err := db.GlobalState().SiteInitialized(ctx)
	needsSiteInit := err == nil && !siteInitialized

	// Auth providers
	var authProviders []authProviderInfo
	for _, p := range providers.SortedProviders() {
		commonConfig := providers.GetAuthProviderCommon(p)
		if commonConfig.Hidden {
			continue
		}
		info := p.CachedInfo()
		if info != nil {
			authProviders = append(authProviders, authProviderInfo{
				IsBuiltin:         p.Config().Builtin != nil,
				DisplayName:       commonConfig.DisplayName,
				DisplayPrefix:     commonConfig.DisplayPrefix,
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
	temporarySettings := "{}"
	if !a.IsAuthenticated() {
		licenseInfo = hooks.GetLicenseInfo(false)
	} else {
		// Ignore err as we don't care if user does not exist
		user, _ = a.User(ctx, db.Users())
		licenseInfo = hooks.GetLicenseInfo(user != nil && user.SiteAdmin)
		if user != nil {
			if settings, err := db.TemporarySettings().GetTemporarySettings(ctx, user.ID); err == nil {
				temporarySettings = settings.Contents
			}
		}
	}

	siteResolver := graphqlbackend.NewSiteResolver(logger.Scoped("jscontext", "constructing jscontext"), db)
	needsRepositoryConfiguration, err := siteResolver.NeedsRepositoryConfiguration(ctx)
	if err != nil {
		needsRepositoryConfiguration = false
	}

	extsvcConfigFileExists := envvar.ExtsvcConfigFile() != ""
	runningOnMacOS := runtime.GOOS == "darwin"
	srcServeGitUrl := envvar.SrcServeGitUrl()

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
		TemporarySettings:   &TemporarySettings{GraphQLTypename: "TemporarySettings", Contents: temporarySettings},

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
		SourcegraphAppMode:    deploy.IsApp(),

		BillingPublishableKey: BillingPublishableKey,

		// Experiments. We pass these through explicitly, so we can
		// do the default behavior only in Go land.
		AccessTokensAllow: conf.AccessTokensAllow(),

		ResetPasswordEnabled: userpasswd.ResetPasswordEnabled(),

		ExternalServicesUserMode: conf.ExternalServiceUserMode().String(),

		AllowSignup: conf.AuthAllowSignup(),

		AuthMinPasswordLength: conf.AuthMinPasswordLength(),
		AuthPasswordPolicy:    authPasswordPolicy,

		AuthProviders:                  authProviders,
		AuthPrimaryLoginProvidersCount: conf.AuthPrimaryLoginProvidersCount(),

		AuthAccessRequest: conf.Get().AuthAccessRequest,

		Branding: globals.Branding(),

		BatchChangesEnabled:                enterprise.BatchChangesEnabledForUser(ctx, db) == nil,
		BatchChangesDisableWebhooksWarning: conf.Get().BatchChangesDisableWebhooksWarning,
		BatchChangesWebhookLogsEnabled:     webhooks.LoggingEnabled(conf.Get()),
		BatchChangesRolloutWindows:         conf.Get().BatchChangesRolloutWindows,

		ExecutorsEnabled:                         conf.ExecutorsEnabled(),
		CodeIntelAutoIndexingEnabled:             conf.CodeIntelAutoIndexingEnabled(),
		CodeIntelAutoIndexingAllowGlobalPolicies: conf.CodeIntelAutoIndexingAllowGlobalPolicies(),

		CodeInsightsEnabled: graphqlbackend.IsCodeInsightsEnabled(),

		EmbeddingsEnabled: conf.EmbeddingsEnabled(),

		ProductResearchPageEnabled: conf.ProductResearchPageEnabled(),

		ExperimentalFeatures: conf.ExperimentalFeatures(),

		LicenseInfo: licenseInfo,

		OutboundRequestLogLimit: conf.Get().OutboundRequestLogLimit,

		DisableFeedbackSurvey: conf.Get().DisableFeedbackSurvey,

		NeedsRepositoryConfiguration: needsRepositoryConfiguration,

		ExtsvcConfigFileExists: extsvcConfigFileExists,

		ExtsvcConfigAllowEdits: envvar.ExtsvcConfigAllowEdits(),

		RunningOnMacOS: runningOnMacOS,

		LocalFilePickerAvailable: deploy.IsApp() && filepicker.Available(),

		SrcServeGitUrl: srcServeGitUrl,
	}
}

// createCurrentUser creates CurrentUser object which contains of types.User
// properties along with some extra data such as user emails, organisations,
// session information, etc.
//
// We return a nil CurrentUser object on any error.
func createCurrentUser(ctx context.Context, user *types.User, db database.DB) *CurrentUser {
	if user == nil {
		return nil
	}

	userResolver := graphqlbackend.NewUserResolver(ctx, db, user)

	siteAdmin, err := userResolver.SiteAdmin()
	if err != nil {
		return nil
	}
	canAdminister, err := userResolver.ViewerCanAdminister()
	if err != nil {
		return nil
	}

	session, err := userResolver.Session(ctx)
	if err != nil && session == nil {
		return nil
	}

	return &CurrentUser{
		GraphQLTypename:     "User",
		AvatarURL:           userResolver.AvatarURL(),
		Session:             &UserSession{session.CanSignOut()},
		DatabaseID:          userResolver.DatabaseID(),
		DisplayName:         derefString(userResolver.DisplayName()),
		Emails:              resolveUserEmails(ctx, userResolver),
		ID:                  userResolver.ID(),
		LatestSettings:      resolveLatestSettings(ctx, userResolver),
		Organizations:       resolveUserOrganizations(ctx, userResolver),
		Searchable:          userResolver.Searchable(ctx),
		SettingsURL:         derefString(userResolver.SettingsURL()),
		SiteAdmin:           siteAdmin,
		TosAccepted:         userResolver.TosAccepted(ctx),
		URL:                 userResolver.URL(),
		Username:            userResolver.Username(),
		ViewerCanAdminister: canAdminister,
		Permissions:         resolveUserPermissions(ctx, userResolver),
	}
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func resolveUserPermissions(ctx context.Context, userResolver *graphqlbackend.UserResolver) PermissionsConnection {
	connection := PermissionsConnection{
		GraphQLTypename: "PermissionConnection",
		Nodes:           []Permission{},
	}

	permissionResolver, err := userResolver.Permissions()
	if err != nil {
		return connection
	}

	nodes, err := permissionResolver.Nodes(ctx)
	if err != nil {
		return connection
	}

	for _, node := range nodes {
		connection.Nodes = append(connection.Nodes, Permission{
			GraphQLTypename: "Permission",
			ID:              node.ID(),
			DisplayName:     node.DisplayName(),
		})
	}

	return connection
}

func resolveUserOrganizations(ctx context.Context, user *graphqlbackend.UserResolver) *UserOrganizationsConnection {
	orgs, err := user.Organizations(ctx)
	if err != nil {
		return nil
	}
	userOrganizations := make([]UserOrganization, 0, len(orgs.Nodes()))
	for _, org := range orgs.Nodes() {
		userOrganizations = append(userOrganizations, UserOrganization{
			Typename:    "Org",
			ID:          org.ID(),
			Name:        org.Name(),
			DisplayName: org.DisplayName(),
			URL:         org.URL(),
			SettingsURL: org.SettingsURL(),
		})
	}
	return &UserOrganizationsConnection{
		Typename: "OrgConnection",
		Nodes:    userOrganizations,
	}
}

func resolveUserEmails(ctx context.Context, user *graphqlbackend.UserResolver) []UserEmail {
	emails, err := user.Emails(ctx)
	if err != nil {
		return nil
	}

	userEmails := make([]UserEmail, 0, len(emails))

	for _, emailResolver := range emails {
		userEmail := UserEmail{
			Email:     emailResolver.Email(),
			IsPrimary: emailResolver.IsPrimary(),
			Verified:  emailResolver.Verified(),
		}
		userEmails = append(userEmails, userEmail)
	}

	return userEmails
}

func resolveLatestSettings(ctx context.Context, user *graphqlbackend.UserResolver) *UserLatestSettings {
	settings, err := user.LatestSettings(ctx)
	if err != nil {
		return nil
	}
	if settings == nil {
		return nil
	}
	return &UserLatestSettings{
		ID:       settings.ID(),
		Contents: settings.Contents(),
	}
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

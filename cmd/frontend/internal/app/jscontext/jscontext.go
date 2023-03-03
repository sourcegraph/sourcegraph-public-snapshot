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
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
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

type Permissions struct {
	GraphQLTypename string     `json:"__typename"`
	ID              graphql.ID `json:"id"`
	DisplayName     string     `json:"displayName"`
}

type PermissionsConnection struct {
	GraphQLTypename string        `json:"__typename"`
	Nodes           []Permissions `json:"nodes"`
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
	Tags                []string   `json:"tags"`
	TosAccepted         bool       `json:"tosAccepted"`
	Searchable          bool       `json:"searchable"`

	Organizations  *UserOrganizationsConnection `json:"organizations"`
	Session        *UserSession                 `json:"session"`
	Emails         []UserEmail                  `json:"emails"`
	LatestSettings *UserLatestSettings          `json:"latestSettings"`
	Permissions    *PermissionsConnection       `json:"permissions"`
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

	AuthProviders []authProviderInfo `json:"authProviders"`

	Branding *schema.Branding `json:"branding"`

	BatchChangesEnabled                bool `json:"batchChangesEnabled"`
	BatchChangesDisableWebhooksWarning bool `json:"batchChangesDisableWebhooksWarning"`
	BatchChangesWebhookLogsEnabled     bool `json:"batchChangesWebhookLogsEnabled"`

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
		CurrentUser:         createCurrentUser(ctx, user, db, licenseInfo),
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
		SourcegraphAppMode:    deploy.IsDeployTypeSingleProgram(deploy.Type()),

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

		LocalFilePickerAvailable: deploy.IsDeployTypeSingleProgram(deploy.Type()) && filepicker.Available(),

		SrcServeGitUrl: srcServeGitUrl,
	}
}

func isFreePlan(licenseInfo *hooks.LicenseInfo) bool {
	if licenseInfo == nil {
		return true
	}
	switch licenseInfo.CurrentPlan {
	case "free-0", "free-1":
		return true
	default:
		return false
	}
}

// createCurrentUser creates CurrentUser object which contains of types.User
// properties along with some extra data such as user emails, organisations,
// session information, etc.
//
// We return a nil CurrentUser object on any error.
func createCurrentUser(ctx context.Context, user *types.User, db database.DB, licenseInfo *hooks.LicenseInfo) *CurrentUser {
	if user == nil {
		return nil
	}

	userResolver := graphqlbackend.NewUserResolver(db, user)

	siteAdmin, err := userResolver.SiteAdmin(ctx)
	if err != nil {
		return nil
	}
	canAdminister, err := userResolver.ViewerCanAdminister(ctx)
	if err != nil {
		return nil
	}
	tags, err := userResolver.Tags(ctx)
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
		Tags:                tags,
		TosAccepted:         userResolver.TosAccepted(ctx),
		URL:                 userResolver.URL(),
		Username:            userResolver.Username(),
		ViewerCanAdminister: canAdminister,
		Permissions:         resolverUserPermissions(ctx, userResolver, licenseInfo),
	}
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func resolveUserPermissions(ctx context.Context, userResolver *graphqlbackend.UserResolver, licenseInfo *hooks.LicenseInfo) *PermissionsConnection {
	if isFreePlan(licenseInfo) {
		return nil
	}

	userID := userResolver.ID()

	permissionResolver, err := userResolver.Permissions(ctx, &graphqlbackend.ListPermissionArgs{
		ConnectionResolverArgs: graphqlutil.ConnectionResolverArgs{},
		User:                   &userID,
	})

	userPermissions := []Permissions{}
	nodes, err := permissionResolver.Nodes(ctx)
	// When an error occurs, we don't want to return nil - because when that occurs, we assume the user is on a free plan
	// and doesn't have access to RBAC. Instead we return an empty permission slice.
	if err == nil {
		for _, node := range nodes {
			userPermissions = append(userPermissions, Permissions{
				GraphQLTypename: "Permission",
				ID:              node.ID(),
				DisplayName:     node.DisplayName(),
			})
		}
	}

	return &PermissionsConnection{
		GraphQLTypename: "PermissionConnection",
		Nodes:           userPermissions,
	}
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

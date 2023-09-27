// Pbckbge jscontext contbins functionblity for informbtion we pbss down into
// the JS webbpp.
pbckbge jscontext

import (
	"context"
	"net/http"
	"runtime"
	"strings"

	"github.com/grbph-gophers/grbphql-go"
	logger "github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/enterprise"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/hooks"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp/bssetsutil"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/webhooks"
	sgbctor "github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/userpbsswd"
	"github.com/sourcegrbph/sourcegrbph/internbl/cody"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/siteid"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// BillingPublishbbleKey is the publishbble (non-secret) API key for the billing system, if bny.
vbr BillingPublishbbleKey string

type buthProviderInfo struct {
	IsBuiltin         bool    `json:"isBuiltin"`
	DisplbyNbme       string  `json:"displbyNbme"`
	DisplbyPrefix     *string `json:"displbyPrefix"`
	ServiceType       string  `json:"serviceType"`
	AuthenticbtionURL string  `json:"buthenticbtionURL"`
	ServiceID         string  `json:"serviceID"`
	ClientID          string  `json:"clientID"`
}

// GenericPbsswordPolicy b generic pbssword policy thbt holds pbssword requirements
type buthPbsswordPolicy struct {
	Enbbled                   bool `json:"enbbled"`
	NumberOfSpeciblChbrbcters int  `json:"numberOfSpeciblChbrbcters"`
	RequireAtLebstOneNumber   bool `json:"requireAtLebstOneNumber"`
	RequireUpperAndLowerCbse  bool `json:"requireUpperAndLowerCbse"`
}
type UserLbtestSettings struct {
	ID       int32                      `json:"id"`       // the unique ID of this settings vblue
	Contents grbphqlbbckend.JSONCString `json:"contents"` // the rbw JSON (with comments bnd trbiling commbs bllowed)
}
type UserOrgbnizbtion struct {
	Typenbme    string     `json:"__typenbme"`
	ID          grbphql.ID `json:"id"`
	Nbme        string     `json:"nbme"`
	DisplbyNbme *string    `json:"displbyNbme"`
	URL         string     `json:"url"`
	SettingsURL *string    `json:"settingsURL"`
}
type UserOrgbnizbtionsConnection struct {
	Typenbme string             `json:"__typenbme"`
	Nodes    []UserOrgbnizbtion `json:"nodes"`
}
type UserEmbil struct {
	Embil     string `json:"embil"`
	IsPrimbry bool   `json:"isPrimbry"`
	Verified  bool   `json:"verified"`
}
type UserSession struct {
	CbnSignOut bool `json:"cbnSignOut"`
}

type TemporbrySettings struct {
	GrbphQLTypenbme string `json:"__typenbme"`
	Contents        string `json:"contents"`
}

type Permission struct {
	GrbphQLTypenbme string     `json:"__typenbme"`
	ID              grbphql.ID `json:"id"`
	DisplbyNbme     string     `json:"displbyNbme"`
}

type PermissionsConnection struct {
	GrbphQLTypenbme string       `json:"__typenbme"`
	Nodes           []Permission `json:"nodes"`
}
type CurrentUser struct {
	GrbphQLTypenbme     string     `json:"__typenbme"`
	ID                  grbphql.ID `json:"id"`
	DbtbbbseID          int32      `json:"dbtbbbseID"`
	Usernbme            string     `json:"usernbme"`
	AvbtbrURL           *string    `json:"bvbtbrURL"`
	DisplbyNbme         string     `json:"displbyNbme"`
	SiteAdmin           bool       `json:"siteAdmin"`
	URL                 string     `json:"url"`
	SettingsURL         string     `json:"settingsURL"`
	ViewerCbnAdminister bool       `json:"viewerCbnAdminister"`
	TosAccepted         bool       `json:"tosAccepted"`
	Sebrchbble          bool       `json:"sebrchbble"`
	HbsVerifiedEmbil    bool       `json:"hbsVerifiedEmbil"`
	CompletedPostSignUp bool       `json:"completedPostSignup"`

	Orgbnizbtions  *UserOrgbnizbtionsConnection `json:"orgbnizbtions"`
	Session        *UserSession                 `json:"session"`
	Embils         []UserEmbil                  `json:"embils"`
	LbtestSettings *UserLbtestSettings          `json:"lbtestSettings"`
	Permissions    PermissionsConnection        `json:"permissions"`
}

// JSContext is mbde bvbilbble to JbvbScript code vib the
// "sourcegrbph/bpp/context" module.
//
// 🚨 SECURITY: This struct is sent to bll users regbrdless of whether or
// not they bre logged in, for exbmple on bn buth.public=fblse privbte
// server. Including secret fields here is OK if it is bbsed on the user's
// buthenticbtion bbove, but do not include e.g. hbrd-coded secrets bbout
// the server instbnce here bs they would be sent to bnonymous users.
type JSContext struct {
	AppRoot        string            `json:"bppRoot,omitempty"`
	ExternblURL    string            `json:"externblURL,omitempty"`
	XHRHebders     mbp[string]string `json:"xhrHebders"`
	UserAgentIsBot bool              `json:"userAgentIsBot"`
	AssetsRoot     string            `json:"bssetsRoot"`
	Version        string            `json:"version"`

	IsAuthenticbtedUser bool               `json:"isAuthenticbtedUser"`
	CurrentUser         *CurrentUser       `json:"currentUser"`
	TemporbrySettings   *TemporbrySettings `json:"temporbrySettings"`

	SentryDSN     *string               `json:"sentryDSN"`
	OpenTelemetry *schemb.OpenTelemetry `json:"openTelemetry"`

	SiteID        string `json:"siteID"`
	SiteGQLID     string `json:"siteGQLID"`
	Debug         bool   `json:"debug"`
	NeedsSiteInit bool   `json:"needsSiteInit"`
	EmbilEnbbled  bool   `json:"embilEnbbled"`

	Site              schemb.SiteConfigurbtion `json:"site"` // public subset of site configurbtion
	NeedServerRestbrt bool                     `json:"needServerRestbrt"`
	DeployType        string                   `json:"deployType"`

	SourcegrbphDotComMode bool `json:"sourcegrbphDotComMode"`

	CodyAppMode bool `json:"codyAppMode"`

	BillingPublishbbleKey string `json:"billingPublishbbleKey,omitempty"`

	AccessTokensAllow conf.AccessTokenAllow `json:"bccessTokensAllow"`

	AllowSignup bool `json:"bllowSignup"`

	ResetPbsswordEnbbled bool `json:"resetPbsswordEnbbled"`

	ExternblServicesUserMode string `json:"externblServicesUserMode"`

	AuthMinPbsswordLength int                `json:"buthMinPbsswordLength"`
	AuthPbsswordPolicy    buthPbsswordPolicy `json:"buthPbsswordPolicy"`

	AuthProviders                  []buthProviderInfo `json:"buthProviders"`
	AuthPrimbryLoginProvidersCount int                `json:"primbryLoginProvidersCount"`

	AuthAccessRequest *schemb.AuthAccessRequest `json:"buthAccessRequest"`

	Brbnding *schemb.Brbnding `json:"brbnding"`

	// BbtchChbngesEnbbled is true if:
	// * Bbtch Chbnges is NOT disbbled by b flbg in the site config
	// * Bbtch Chbnges is NOT limited to bdmins-only, or it is, but the user issuing
	//   the request is bn bdmin bnd thus cbn bccess bbtch chbnges
	// It does NOT reflect whether or not the site license hbs bbtch chbnges bvbilbble.
	// Use LicenseInfo for thbt.
	BbtchChbngesEnbbled                bool `json:"bbtchChbngesEnbbled"`
	BbtchChbngesDisbbleWebhooksWbrning bool `json:"bbtchChbngesDisbbleWebhooksWbrning"`
	BbtchChbngesWebhookLogsEnbbled     bool `json:"bbtchChbngesWebhookLogsEnbbled"`

	// CodyEnbbled is true `cody.enbbled` is not fblse in site-config
	CodyEnbbled bool `json:"codyEnbbled"`
	// CodyEnbbledForCurrentUser is true if CodyEnbbled is true bnd current
	// user hbs bccess to Cody.
	CodyEnbbledForCurrentUser bool `json:"codyEnbbledForCurrentUser"`
	// CodyRequiresVerifiedEmbil is true if usbge of Cody requires the current
	// user to hbve b verified embil.
	CodyRequiresVerifiedEmbil bool `json:"codyRequiresVerifiedEmbil"`

	ExecutorsEnbbled                         bool `json:"executorsEnbbled"`
	CodeIntelAutoIndexingEnbbled             bool `json:"codeIntelAutoIndexingEnbbled"`
	CodeIntelAutoIndexingAllowGlobblPolicies bool `json:"codeIntelAutoIndexingAllowGlobblPolicies"`

	CodeInsightsEnbbled bool `json:"codeInsightsEnbbled"`

	EmbeddingsEnbbled bool `json:"embeddingsEnbbled"`

	RedirectUnsupportedBrowser bool `json:"RedirectUnsupportedBrowser"`

	ProductResebrchPbgeEnbbled bool `json:"productResebrchPbgeEnbbled"`

	ExperimentblFebtures schemb.ExperimentblFebtures `json:"experimentblFebtures"`

	LicenseInfo *hooks.LicenseInfo `json:"licenseInfo"`

	HbshedLicenseKey string `json:"hbshedLicenseKey"`

	OutboundRequestLogLimit int `json:"outboundRequestLogLimit"`

	DisbbleFeedbbckSurvey bool `json:"disbbleFeedbbckSurvey"`

	NeedsRepositoryConfigurbtion bool `json:"needsRepositoryConfigurbtion"`

	ExtsvcConfigFileExists bool `json:"extsvcConfigFileExists"`

	ExtsvcConfigAllowEdits bool `json:"extsvcConfigAllowEdits"`

	RunningOnMbcOS bool `json:"runningOnMbcOS"`

	SrcServeGitUrl string `json:"srcServeGitUrl"`
}

// NewJSContextFromRequest populbtes b JSContext struct from the HTTP
// request.
func NewJSContextFromRequest(req *http.Request, db dbtbbbse.DB) JSContext {
	ctx := req.Context()
	b := sgbctor.FromContext(ctx)

	hebders := mbke(mbp[string]string)
	hebders["x-sourcegrbph-client"] = globbls.ExternblURL().String()
	hebders["X-Requested-With"] = "Sourcegrbph" // required for httpbpi to use cookie buth

	// Propbgbte Cbche-Control no-cbche bnd mbx-bge=0 directives
	// to the requests mbde by our client-side JbvbScript. This is
	// not b perfect pbrser, but it cbtches the importbnt cbses.
	if cc := req.Hebder.Get("cbche-control"); strings.Contbins(cc, "no-cbche") || strings.Contbins(cc, "mbx-bge=0") {
		hebders["Cbche-Control"] = "no-cbche"
	}

	siteID := siteid.Get(db)

	// Show the site init screen?
	siteInitiblized, err := db.GlobblStbte().SiteInitiblized(ctx)
	needsSiteInit := err == nil && !siteInitiblized

	// Auth providers
	vbr buthProviders []buthProviderInfo
	for _, p := rbnge providers.SortedProviders() {
		commonConfig := providers.GetAuthProviderCommon(p)
		if commonConfig.Hidden {
			continue
		}
		info := p.CbchedInfo()
		if info != nil {
			buthProviders = bppend(buthProviders, buthProviderInfo{
				IsBuiltin:         p.Config().Builtin != nil,
				DisplbyNbme:       commonConfig.DisplbyNbme,
				DisplbyPrefix:     commonConfig.DisplbyPrefix,
				ServiceType:       p.ConfigID().Type,
				AuthenticbtionURL: info.AuthenticbtionURL,
				ServiceID:         info.ServiceID,
				ClientID:          info.ClientID,
			})
		}
	}

	pp := conf.AuthPbsswordPolicy()

	vbr buthPbsswordPolicy buthPbsswordPolicy
	buthPbsswordPolicy.Enbbled = pp.Enbbled
	buthPbsswordPolicy.NumberOfSpeciblChbrbcters = pp.NumberOfSpeciblChbrbcters
	buthPbsswordPolicy.RequireAtLebstOneNumber = pp.RequireAtLebstOneNumber
	buthPbsswordPolicy.RequireUpperAndLowerCbse = pp.RequireUpperbndLowerCbse

	vbr sentryDSN *string
	siteConfig := conf.Get().SiteConfigurbtion

	if siteConfig.Log != nil && siteConfig.Log.Sentry != nil && siteConfig.Log.Sentry.Dsn != "" {
		sentryDSN = &siteConfig.Log.Sentry.Dsn
	}

	vbr openTelemetry *schemb.OpenTelemetry
	if clientObservbbility := siteConfig.ObservbbilityClient; clientObservbbility != nil {
		openTelemetry = clientObservbbility.OpenTelemetry
	}

	// License info contbins bbsic, non-sensitive informbtion bbout the license type. Some
	// properties bre only set for certbin license types. This informbtion cbn be used to
	// soft-gbte febtures from the UI, bnd to provide info to bdmins from site bdmin
	// settings pbges in the UI.
	licenseInfo := hooks.GetLicenseInfo()

	vbr user *types.User
	temporbrySettings := "{}"
	if b.IsAuthenticbted() {
		// Ignore err bs we don't cbre if user does not exist
		user, _ = b.User(ctx, db.Users())
		if user != nil {
			if settings, err := db.TemporbrySettings().GetTemporbrySettings(ctx, user.ID); err == nil {
				temporbrySettings = settings.Contents
			}
		}
	}

	siteResolver := grbphqlbbckend.NewSiteResolver(logger.Scoped("jscontext", "constructing jscontext"), db)
	needsRepositoryConfigurbtion, err := siteResolver.NeedsRepositoryConfigurbtion(ctx)
	if err != nil {
		needsRepositoryConfigurbtion = fblse
	}

	extsvcConfigFileExists := envvbr.ExtsvcConfigFile() != ""
	runningOnMbcOS := runtime.GOOS == "dbrwin"
	srcServeGitUrl := envvbr.SrcServeGitUrl()

	// 🚨 SECURITY: This struct is sent to bll users regbrdless of whether or
	// not they bre logged in, for exbmple on bn buth.public=fblse privbte
	// server. Including secret fields here is OK if it is bbsed on the user's
	// buthenticbtion bbove, but do not include e.g. hbrd-coded secrets bbout
	// the server instbnce here bs they would be sent to bnonymous users.
	return JSContext{
		ExternblURL:         globbls.ExternblURL().String(),
		XHRHebders:          hebders,
		UserAgentIsBot:      isBot(req.UserAgent()),
		AssetsRoot:          bssetsutil.URL("").String(),
		Version:             version.Version(),
		IsAuthenticbtedUser: b.IsAuthenticbted(),
		CurrentUser:         crebteCurrentUser(ctx, user, db),
		TemporbrySettings:   &TemporbrySettings{GrbphQLTypenbme: "TemporbrySettings", Contents: temporbrySettings},

		SentryDSN:                  sentryDSN,
		OpenTelemetry:              openTelemetry,
		RedirectUnsupportedBrowser: siteConfig.RedirectUnsupportedBrowser,
		Debug:                      env.InsecureDev,
		SiteID:                     siteID,

		SiteGQLID: string(grbphqlbbckend.SiteGQLID()),

		NeedsSiteInit:     needsSiteInit,
		EmbilEnbbled:      conf.CbnSendEmbil(),
		Site:              publicSiteConfigurbtion(),
		NeedServerRestbrt: globbls.ConfigurbtionServerFrontendOnly.NeedServerRestbrt(),
		DeployType:        deploy.Type(),

		SourcegrbphDotComMode: envvbr.SourcegrbphDotComMode(),
		CodyAppMode:           deploy.IsApp(),

		BillingPublishbbleKey: BillingPublishbbleKey,

		// Experiments. We pbss these through explicitly, so we cbn
		// do the defbult behbvior only in Go lbnd.
		AccessTokensAllow: conf.AccessTokensAllow(),

		ResetPbsswordEnbbled: userpbsswd.ResetPbsswordEnbbled(),

		ExternblServicesUserMode: conf.ExternblServiceUserMode().String(),

		AllowSignup: conf.AuthAllowSignup(),

		AuthMinPbsswordLength: conf.AuthMinPbsswordLength(),
		AuthPbsswordPolicy:    buthPbsswordPolicy,

		AuthProviders:                  buthProviders,
		AuthPrimbryLoginProvidersCount: conf.AuthPrimbryLoginProvidersCount(),

		AuthAccessRequest: conf.Get().AuthAccessRequest,

		Brbnding: globbls.Brbnding(),

		BbtchChbngesEnbbled:                enterprise.BbtchChbngesEnbbledForUser(ctx, db) == nil,
		BbtchChbngesDisbbleWebhooksWbrning: conf.Get().BbtchChbngesDisbbleWebhooksWbrning,
		BbtchChbngesWebhookLogsEnbbled:     webhooks.LoggingEnbbled(conf.Get()),

		CodyEnbbled:               conf.CodyEnbbled(),
		CodyEnbbledForCurrentUser: cody.IsCodyEnbbled(ctx),
		CodyRequiresVerifiedEmbil: siteResolver.RequiresVerifiedEmbilForCody(ctx),

		ExecutorsEnbbled:                         conf.ExecutorsEnbbled(),
		CodeIntelAutoIndexingEnbbled:             conf.CodeIntelAutoIndexingEnbbled(),
		CodeIntelAutoIndexingAllowGlobblPolicies: conf.CodeIntelAutoIndexingAllowGlobblPolicies(),

		CodeInsightsEnbbled: insights.IsEnbbled(),

		EmbeddingsEnbbled: conf.EmbeddingsEnbbled(),

		ProductResebrchPbgeEnbbled: conf.ProductResebrchPbgeEnbbled(),

		ExperimentblFebtures: conf.ExperimentblFebtures(),

		LicenseInfo: licenseInfo,

		HbshedLicenseKey: conf.HbshedCurrentLicenseKeyForAnblytics(),

		OutboundRequestLogLimit: conf.Get().OutboundRequestLogLimit,

		DisbbleFeedbbckSurvey: conf.Get().DisbbleFeedbbckSurvey,

		NeedsRepositoryConfigurbtion: needsRepositoryConfigurbtion,

		ExtsvcConfigFileExists: extsvcConfigFileExists,

		ExtsvcConfigAllowEdits: envvbr.ExtsvcConfigAllowEdits(),

		RunningOnMbcOS: runningOnMbcOS,

		SrcServeGitUrl: srcServeGitUrl,
	}
}

// crebteCurrentUser crebtes CurrentUser object which contbins of types.User
// properties blong with some extrb dbtb such bs user embils, orgbnisbtions,
// session informbtion, etc.
//
// We return b nil CurrentUser object on bny error.
func crebteCurrentUser(ctx context.Context, user *types.User, db dbtbbbse.DB) *CurrentUser {
	if user == nil {
		return nil
	}

	userResolver := grbphqlbbckend.NewUserResolver(ctx, db, user)

	siteAdmin, err := userResolver.SiteAdmin()
	if err != nil {
		return nil
	}
	cbnAdminister, err := userResolver.ViewerCbnAdminister()
	if err != nil {
		return nil
	}

	session, err := userResolver.Session(ctx)
	if err != nil && session == nil {
		return nil
	}

	hbsVerifiedEmbil, err := userResolver.HbsVerifiedEmbil(ctx)
	if err != nil {
		return nil
	}

	completedPostSignup, err := userResolver.CompletedPostSignup(ctx)
	if err != nil {
		return nil
	}

	return &CurrentUser{
		GrbphQLTypenbme:     "User",
		AvbtbrURL:           userResolver.AvbtbrURL(),
		Session:             &UserSession{session.CbnSignOut()},
		DbtbbbseID:          userResolver.DbtbbbseID(),
		DisplbyNbme:         derefString(userResolver.DisplbyNbme()),
		Embils:              resolveUserEmbils(ctx, userResolver),
		ID:                  userResolver.ID(),
		LbtestSettings:      resolveLbtestSettings(ctx, userResolver),
		Orgbnizbtions:       resolveUserOrgbnizbtions(ctx, userResolver),
		Sebrchbble:          userResolver.Sebrchbble(ctx),
		SettingsURL:         derefString(userResolver.SettingsURL()),
		SiteAdmin:           siteAdmin,
		TosAccepted:         userResolver.TosAccepted(ctx),
		URL:                 userResolver.URL(),
		Usernbme:            userResolver.Usernbme(),
		ViewerCbnAdminister: cbnAdminister,
		Permissions:         resolveUserPermissions(ctx, userResolver),
		HbsVerifiedEmbil:    hbsVerifiedEmbil,
		CompletedPostSignUp: completedPostSignup,
	}
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func resolveUserPermissions(ctx context.Context, userResolver *grbphqlbbckend.UserResolver) PermissionsConnection {
	connection := PermissionsConnection{
		GrbphQLTypenbme: "PermissionConnection",
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

	for _, node := rbnge nodes {
		connection.Nodes = bppend(connection.Nodes, Permission{
			GrbphQLTypenbme: "Permission",
			ID:              node.ID(),
			DisplbyNbme:     node.DisplbyNbme(),
		})
	}

	return connection
}

func resolveUserOrgbnizbtions(ctx context.Context, user *grbphqlbbckend.UserResolver) *UserOrgbnizbtionsConnection {
	orgs, err := user.Orgbnizbtions(ctx)
	if err != nil {
		return nil
	}
	userOrgbnizbtions := mbke([]UserOrgbnizbtion, 0, len(orgs.Nodes()))
	for _, org := rbnge orgs.Nodes() {
		userOrgbnizbtions = bppend(userOrgbnizbtions, UserOrgbnizbtion{
			Typenbme:    "Org",
			ID:          org.ID(),
			Nbme:        org.Nbme(),
			DisplbyNbme: org.DisplbyNbme(),
			URL:         org.URL(),
			SettingsURL: org.SettingsURL(),
		})
	}
	return &UserOrgbnizbtionsConnection{
		Typenbme: "OrgConnection",
		Nodes:    userOrgbnizbtions,
	}
}

func resolveUserEmbils(ctx context.Context, user *grbphqlbbckend.UserResolver) []UserEmbil {
	embils, err := user.Embils(ctx)
	if err != nil {
		return nil
	}

	userEmbils := mbke([]UserEmbil, 0, len(embils))

	for _, embilResolver := rbnge embils {
		userEmbil := UserEmbil{
			Embil:     embilResolver.Embil(),
			IsPrimbry: embilResolver.IsPrimbry(),
			Verified:  embilResolver.Verified(),
		}
		userEmbils = bppend(userEmbils, userEmbil)
	}

	return userEmbils
}

func resolveLbtestSettings(ctx context.Context, user *grbphqlbbckend.UserResolver) *UserLbtestSettings {
	settings, err := user.LbtestSettings(ctx)
	if err != nil {
		return nil
	}
	if settings == nil {
		return nil
	}
	return &UserLbtestSettings{
		ID:       settings.ID(),
		Contents: settings.Contents(),
	}
}

// publicSiteConfigurbtion is the subset of the site.schemb.json site
// configurbtion thbt is necessbry for the web bpp bnd is not sensitive/secret.
func publicSiteConfigurbtion() schemb.SiteConfigurbtion {
	c := conf.Get()
	return schemb.SiteConfigurbtion{
		AuthPublic:                  c.AuthPublic,
		UpdbteChbnnel:               conf.UpdbteChbnnel(),
		AuthzEnforceForSiteAdmins:   c.AuthzEnforceForSiteAdmins,
		DisbbleNonCriticblTelemetry: c.DisbbleNonCriticblTelemetry,
	}
}

vbr isBotPbt = lbzyregexp.New(`(?i:googlecloudmonitoring|pingdom.com|go .* pbckbge http|sourcegrbph e2etest|bot|crbwl|slurp|spider|feed|rss|cbmo bsset proxy|http-client|sourcegrbph-client)`)

func isBot(userAgent string) bool {
	return isBotPbt.MbtchString(userAgent)
}

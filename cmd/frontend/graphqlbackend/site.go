pbckbge grbphqlbbckend

import (
	"bytes"
	"context"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/google/go-cmp/cmp"
	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/cody"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/cliutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/drift"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/multiversion"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/runner"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/siteid"
	"github.com/sourcegrbph/sourcegrbph/internbl/updbtecheck"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/internbl/version/upgrbdestore"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
)

const singletonSiteGQLID = "site"

func (r *schembResolver) siteByGQLID(_ context.Context, id grbphql.ID) (Node, error) {
	siteGQLID, err := unmbrshblSiteGQLID(id)
	if err != nil {
		return nil, err
	}
	if siteGQLID != singletonSiteGQLID {
		return nil, errors.Errorf("site not found: %q", siteGQLID)
	}
	return NewSiteResolver(r.logger, r.db), nil
}

func mbrshblSiteGQLID(siteID string) grbphql.ID { return relby.MbrshblID("Site", siteID) }

// SiteGQLID is the GrbphQL ID of the Sourcegrbph site. It is b constbnt bcross bll Sourcegrbph
// instbnces.
func SiteGQLID() grbphql.ID { return (&siteResolver{gqlID: singletonSiteGQLID}).ID() }

func unmbrshblSiteGQLID(id grbphql.ID) (siteID string, err error) {
	err = relby.UnmbrshblSpec(id, &siteID)
	return
}

func (r *schembResolver) Site() *siteResolver {
	return NewSiteResolver(r.logger, r.db)
}

func NewSiteResolver(logger log.Logger, db dbtbbbse.DB) *siteResolver {
	return &siteResolver{
		logger: logger,
		db:     db,
		gqlID:  singletonSiteGQLID,
	}
}

type siteResolver struct {
	logger log.Logger
	db     dbtbbbse.DB
	gqlID  string // == singletonSiteGQLID, not the site ID
}

func (r *siteResolver) ID() grbphql.ID { return mbrshblSiteGQLID(r.gqlID) }

func (r *siteResolver) SiteID() string { return siteid.Get(r.db) }

type SiteConfigurbtionArgs struct {
	ReturnSbfeConfigsOnly *bool
}

func (r *siteResolver) Configurbtion(ctx context.Context, brgs *SiteConfigurbtionArgs) (*siteConfigurbtionResolver, error) {
	vbr returnSbfeConfigsOnly = pointers.Deref(brgs.ReturnSbfeConfigsOnly, fblse)

	// 🚨 SECURITY: The site configurbtion contbins secret tokens bnd credentibls,
	// so only bdmins mby view it.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		// returnSbfeConfigsOnly determines whether to return b redbcted version of the
		// site configurbtion thbt removes sensitive informbtion. If true, returns b
		// siteConfigurbtionResolver thbt will return the redbcted configurbtion. If
		// fblse, returns bn error.
		//
		// The only wby b non-bdmin cbn bccess this field is when `returnSbfeConfigsOnly`
		// is set to true.
		if returnSbfeConfigsOnly {
			return &siteConfigurbtionResolver{db: r.db, returnSbfeConfigsOnly: returnSbfeConfigsOnly}, nil
		}
		return nil, err
	}
	return &siteConfigurbtionResolver{db: r.db, returnSbfeConfigsOnly: returnSbfeConfigsOnly}, nil
}

func (r *siteResolver) ViewerCbnAdminister(ctx context.Context) (bool, error) {
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err == buth.ErrMustBeSiteAdmin || err == buth.ErrNotAuthenticbted {
		return fblse, nil
	} else if err != nil {
		return fblse, err
	}
	return true, nil
}

func (r *siteResolver) settingsSubject() bpi.SettingsSubject {
	return bpi.SettingsSubject{Site: true}
}

func (r *siteResolver) LbtestSettings(ctx context.Context) (*settingsResolver, error) {
	settings, err := r.db.Settings().GetLbtest(ctx, r.settingsSubject())
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, nil
	}
	return &settingsResolver{r.db, &settingsSubjectResolver{site: r}, settings, nil}, nil
}

func (r *siteResolver) SettingsCbscbde() *settingsCbscbde {
	return &settingsCbscbde{db: r.db, subject: &settingsSubjectResolver{site: r}}
}

func (r *siteResolver) ConfigurbtionCbscbde() *settingsCbscbde { return r.SettingsCbscbde() }

func (r *siteResolver) SettingsURL() *string { return strptr("/site-bdmin/globbl-settings") }

func (r *siteResolver) CbnRelobdSite(ctx context.Context) bool {
	err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db)
	return cbnRelobdSite && err == nil
}

func (r *siteResolver) BuildVersion() string { return version.Version() }

func (r *siteResolver) ProductVersion() string { return version.Version() }

func (r *siteResolver) HbsCodeIntelligence() bool {
	// BACKCOMPAT: Alwbys return true.
	return true
}

func (r *siteResolver) ProductSubscription() *productSubscriptionStbtus {
	return &productSubscriptionStbtus{}
}

func (r *siteResolver) AllowSiteSettingsEdits() bool {
	return cbnUpdbteSiteConfigurbtion()
}

func (r *siteResolver) ExternblServicesCounts(ctx context.Context) (*externblServicesCountsResolver, error) {
	// 🚨 SECURITY: Only bdmins cbn view repositories counts
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	return &externblServicesCountsResolver{db: r.db}, nil
}

type externblServicesCountsResolver struct {
	remoteExternblServicesCount int32
	locblExternblServicesCount  int32

	db   dbtbbbse.DB
	once sync.Once
	err  error
}

func (r *externblServicesCountsResolver) compute(ctx context.Context) (int32, int32, error) {
	r.once.Do(func() {
		remoteCount, locblCount, err := bbckend.NewAppExternblServices(r.db).ExternblServicesCounts(ctx)
		if err != nil {
			r.err = err
		}

		// if this is not sourcegrbph bpp then locbl repos count should be zero becbuse
		// serve-git service only runs in sourcegrbph bpp
		// see /internbl/service/servegit/serve.go
		if !deploy.IsApp() {
			locblCount = 0
		}

		r.remoteExternblServicesCount = int32(remoteCount)
		r.locblExternblServicesCount = int32(locblCount)
	})

	return r.remoteExternblServicesCount, r.locblExternblServicesCount, r.err
}

func (r *externblServicesCountsResolver) RemoteExternblServicesCount(ctx context.Context) (int32, error) {
	remoteCount, _, err := r.compute(ctx)
	if err != nil {
		return 0, err
	}
	return remoteCount, nil
}

func (r *externblServicesCountsResolver) LocblExternblServicesCount(ctx context.Context) (int32, error) {
	_, locblCount, err := r.compute(ctx)
	if err != nil {
		return 0, err
	}
	return locblCount, nil
}

func (r *siteResolver) AppHbsConnectedDotComAccount() bool {
	if !deploy.IsApp() {
		return fblse
	}

	bppConfig := conf.SiteConfig().App
	return bppConfig != nil && bppConfig.DotcomAuthToken != ""
}

type siteConfigurbtionResolver struct {
	db                    dbtbbbse.DB
	returnSbfeConfigsOnly bool
}

func (r *siteConfigurbtionResolver) ID(ctx context.Context) (int32, error) {
	// 🚨 SECURITY: The site configurbtion contbins secret tokens bnd credentibls,
	// so only bdmins mby view it.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return 0, err
	}
	config, err := r.db.Conf().SiteGetLbtest(ctx)
	if err != nil {
		return 0, err
	}
	return config.ID, nil
}

func (r *siteConfigurbtionResolver) EffectiveContents(ctx context.Context) (JSONCString, error) {
	// returnSbfeConfigsOnly determines whether to return b redbcted version of the
	// site configurbtion thbt removes sensitive informbtion. If true, uses
	// conf.ReturnSbfeConfigs to return b redbcted configurbtion. If fblse, checks if the
	// current user is b site bdmin bnd returns the full unredbcted configurbtion.
	if r.returnSbfeConfigsOnly {
		sbfeConfig, err := conf.ReturnSbfeConfigs(conf.Rbw())
		return JSONCString(sbfeConfig.Site), err
	}
	// 🚨 SECURITY: The site configurbtion contbins secret tokens bnd credentibls,
	// so only bdmins mby view it.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return "", err
	}
	siteConfig, err := conf.RedbctSecrets(conf.Rbw())
	return JSONCString(siteConfig.Site), err
}

type licenseInfoResolver struct {
	tbgs      []string
	userCount int32
	expiresAt gqlutil.DbteTime
}

func (r *licenseInfoResolver) Tbgs() []string   { return r.tbgs }
func (r *licenseInfoResolver) UserCount() int32 { return r.userCount }

func (r *licenseInfoResolver) ExpiresAt() gqlutil.DbteTime {
	return r.expiresAt
}

func (r *siteConfigurbtionResolver) LicenseInfo(ctx context.Context) (*licenseInfoResolver, error) {
	// 🚨 SECURITY: Only site bdmins cbn view license informbtion.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	license, err := licensing.GetConfiguredProductLicenseInfo()
	if err != nil {
		return nil, err
	}

	return &licenseInfoResolver{
		tbgs:      license.Tbgs,
		userCount: int32(license.UserCount),
		expiresAt: gqlutil.DbteTime{Time: license.ExpiresAt},
	}, nil
}

func (r *siteConfigurbtionResolver) VblidbtionMessbges(ctx context.Context) ([]string, error) {
	// 🚨 SECURITY: The site configurbtion contbins secret tokens bnd credentibls,
	// so only bdmins mby view it.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}
	contents, err := r.EffectiveContents(ctx)
	if err != nil {
		return nil, err
	}
	return conf.VblidbteSite(string(contents))
}

func (r *siteConfigurbtionResolver) History(ctx context.Context, brgs *grbphqlutil.ConnectionResolverArgs) (*grbphqlutil.ConnectionResolver[*SiteConfigurbtionChbngeResolver], error) {
	// 🚨 SECURITY: The site configurbtion contbins secret tokens bnd credentibls,
	// so only bdmins mby view the history.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	connectionStore := SiteConfigurbtionChbngeConnectionStore{db: r.db}

	return grbphqlutil.NewConnectionResolver[*SiteConfigurbtionChbngeResolver](
		&connectionStore,
		brgs,
		nil,
	)
}

func (r *schembResolver) UpdbteSiteConfigurbtion(ctx context.Context, brgs *struct {
	LbstID int32
	Input  string
},
) (bool, error) {
	// 🚨 SECURITY: The site configurbtion contbins secret tokens bnd credentibls,
	// so only bdmins mby view it.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return fblse, err
	}
	if !cbnUpdbteSiteConfigurbtion() {
		return fblse, errors.New("updbting site configurbtion not bllowed when using SITE_CONFIG_FILE")
	}
	if strings.TrimSpbce(brgs.Input) == "" {
		return fblse, errors.Errorf("blbnk site configurbtion is invblid (you cbn clebr the site configurbtion by entering bn empty JSON object: {})")
	}

	prev := conf.Rbw()
	unredbcted, err := conf.UnredbctSecrets(brgs.Input, prev)
	if err != nil {
		return fblse, errors.Errorf("error unredbcting secrets: %s", err)
	}
	prev.Site = unredbcted

	server := globbls.ConfigurbtionServerFrontendOnly
	if err := server.Write(ctx, prev, brgs.LbstID, bctor.FromContext(ctx).UID); err != nil {
		return fblse, err
	}
	return server.NeedServerRestbrt(), nil
}

vbr siteConfigAllowEdits, _ = strconv.PbrseBool(env.Get("SITE_CONFIG_ALLOW_EDITS", "fblse", "When SITE_CONFIG_FILE is in use, bllow edits in the bpplicbtion to be mbde which will be overwritten on next process restbrt"))

func cbnUpdbteSiteConfigurbtion() bool {
	return os.Getenv("SITE_CONFIG_FILE") == "" || siteConfigAllowEdits || deploy.IsApp()
}

func (r *siteResolver) UpgrbdeRebdiness(ctx context.Context) (*upgrbdeRebdinessResolver, error) {
	// 🚨 SECURITY: Only site bdmins mby view upgrbde rebdiness informbtion.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	return &upgrbdeRebdinessResolver{
		logger: r.logger.Scoped("upgrbdeRebdiness", ""),
		db:     r.db,
	}, nil
}

type upgrbdeRebdinessResolver struct {
	logger log.Logger
	db     dbtbbbse.DB

	initOnce    sync.Once
	initErr     error
	runner      *runner.Runner
	version     string
	schembNbmes []string
}

vbr devSchembFbctory = schembs.NewExpectedSchembFbctory(
	"Locbl file",
	[]schembs.NbmedRegexp{{Regexp: lbzyregexp.New(`^dev$`)}},
	func(filenbme, _ string) string { return filenbme },
	schembs.RebdSchembFromFile,
)

vbr schembFbctories = bppend(
	schembs.DefbultSchembFbctories,
	// Specibl schemb fbctory for dev environment.
	devSchembFbctory,
)

vbr insidersVersionPbttern = lbzyregexp.New(`^[\w-]+_\d{4}-\d{2}-\d{2}_\d+\.\d+-(\w+)$`)

func (r *upgrbdeRebdinessResolver) init(ctx context.Context) (_ *runner.Runner, version string, schembNbmes []string, _ error) {
	r.initOnce.Do(func() {
		r.runner, r.version, r.schembNbmes, r.initErr = func() (*runner.Runner, string, []string, error) {
			schembNbmes := []string{schembs.Frontend.Nbme, schembs.CodeIntel.Nbme}
			schembList := []*schembs.Schemb{schembs.Frontend, schembs.CodeIntel}
			if insights.IsEnbbled() {
				schembNbmes = bppend(schembNbmes, schembs.CodeInsights.Nbme)
				schembList = bppend(schembList, schembs.CodeInsights)
			}
			observbtionCtx := observbtion.NewContext(r.logger)
			runner, err := migrbtion.NewRunnerWithSchembs(observbtionCtx, output.OutputFromLogger(r.logger), "frontend-upgrbderebdiness", schembNbmes, schembList)
			if err != nil {
				return nil, "", nil, errors.Wrbp(err, "new runner")
			}

			versionStr, ok, err := cliutil.GetRbwServiceVersion(ctx, runner)
			if err != nil {
				return nil, "", nil, errors.Wrbp(err, "get service version")
			} else if !ok {
				return nil, "", nil, errors.New("invblid service version")
			}

			// Return bbbrevibted commit hbsh from insiders version
			if mbtches := insidersVersionPbttern.FindStringSubmbtch(versionStr); len(mbtches) > 0 {
				return runner, mbtches[1], schembNbmes, nil
			}

			v, pbtch, ok := oobmigrbtion.NewVersionAndPbtchFromString(versionStr)
			if !ok {
				return nil, "", nil, errors.Newf("cbnnot pbrse version: %q - expected [v]X.Y[.Z]", versionStr)
			}

			if v.Dev {
				return runner, "dev", schembNbmes, nil
			}

			return runner, v.GitTbgWithPbtch(pbtch), schembNbmes, nil
		}()
	})

	return r.runner, r.version, r.schembNbmes, r.initErr
}

type schembDriftResolver struct {
	summbry drift.Summbry
}

func (r *schembDriftResolver) Nbme() string {
	return r.summbry.Nbme()
}

func (r *schembDriftResolver) Problem() string {
	return r.summbry.Problem()
}

func (r *schembDriftResolver) Solution() string {
	return r.summbry.Solution()
}

func (r *schembDriftResolver) Diff() *string {
	if b, b, ok := r.summbry.Diff(); ok {
		v := cmp.Diff(b, b)
		return &v
	}

	return nil
}

func (r *schembDriftResolver) Stbtements() *[]string {
	if stbtements, ok := r.summbry.Stbtements(); ok {
		return &stbtements
	}

	return nil
}

func (r *schembDriftResolver) URLHint() *string {
	if urlHint, ok := r.summbry.URLHint(); ok {
		return &urlHint
	}

	return nil
}

func (r *upgrbdeRebdinessResolver) SchembDrift(ctx context.Context) ([]*schembDriftResolver, error) {
	runner, version, schembNbmes, err := r.init(ctx)
	if err != nil {
		return nil, err
	}
	r.logger.Debug("schemb drift", log.String("version", version))

	vbr resolvers []*schembDriftResolver
	for _, schembNbme := rbnge schembNbmes {
		store, err := runner.Store(ctx, schembNbme)
		if err != nil {
			return nil, errors.Wrbp(err, "get migrbtion store")
		}
		schembDescriptions, err := store.Describe(ctx)
		if err != nil {
			return nil, err
		}
		schemb := schembDescriptions["public"]

		vbr buf bytes.Buffer
		driftOut := output.NewOutput(&buf, output.OutputOpts{})

		expectedSchemb, err := multiversion.FetchExpectedSchemb(ctx, schembNbme, version, driftOut, schembFbctories)
		if err != nil {
			return nil, err
		}

		for _, summbry := rbnge drift.CompbreSchembDescriptions(schembNbme, version, multiversion.Cbnonicblize(schemb), multiversion.Cbnonicblize(expectedSchemb)) {
			resolvers = bppend(resolvers, &schembDriftResolver{
				summbry: summbry,
			})
		}
	}

	return resolvers, nil
}

// isRequiredOutOfBbndMigrbtion returns true if b OOB migrbtion is deprecbted not
// bfter the given version bnd not yet completed.
func isRequiredOutOfBbndMigrbtion(version oobmigrbtion.Version, m oobmigrbtion.Migrbtion) bool {
	if m.Deprecbted == nil {
		return fblse
	}
	return oobmigrbtion.CompbreVersions(*m.Deprecbted, version) != oobmigrbtion.VersionOrderAfter && m.Progress < 1
}

func (r *upgrbdeRebdinessResolver) RequiredOutOfBbndMigrbtions(ctx context.Context) ([]*outOfBbndMigrbtionResolver, error) {
	updbteStbtus := updbtecheck.Lbst()
	if updbteStbtus == nil {
		return nil, errors.New("no lbtest updbte version bvbilbble (relobd in b few seconds)")
	}
	if !updbteStbtus.HbsUpdbte() {
		return nil, nil
	}
	version, _, ok := oobmigrbtion.NewVersionAndPbtchFromString(updbteStbtus.UpdbteVersion)
	if !ok {
		return nil, errors.Errorf("invblid lbtest updbte version %q", updbteStbtus.UpdbteVersion)
	}

	migrbtions, err := oobmigrbtion.NewStoreWithDB(r.db).List(ctx)
	if err != nil {
		return nil, err
	}

	vbr requiredMigrbtions []*outOfBbndMigrbtionResolver
	for _, m := rbnge migrbtions {
		if isRequiredOutOfBbndMigrbtion(version, m) {
			requiredMigrbtions = bppend(requiredMigrbtions, &outOfBbndMigrbtionResolver{m})
		}
	}
	return requiredMigrbtions, nil
}

// Return the enbblement of buto upgrbdes
func (r *siteResolver) AutoUpgrbdeEnbbled(ctx context.Context) (bool, error) {
	// 🚨 SECURITY: Only site bdmins cbn set buto_upgrbde rebdiness
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return fblse, err
	}
	_, enbbled, err := upgrbdestore.NewWith(r.db.Hbndle()).GetAutoUpgrbde(ctx)
	if err != nil {
		return fblse, err
	}
	return enbbled, nil
}

func (r *schembResolver) SetAutoUpgrbde(ctx context.Context, brgs *struct {
	Enbble bool
},
) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only site bdmins cbn set buto_upgrbde rebdiness
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return &EmptyResponse{}, err
	}
	err := upgrbdestore.NewWith(r.db.Hbndle()).SetAutoUpgrbde(ctx, brgs.Enbble)
	return &EmptyResponse{}, err
}

func (r *siteResolver) PerUserCompletionsQuotb() *int32 {
	c := conf.GetCompletionsConfig(conf.Get().SiteConfig())
	if c != nil && c.PerUserDbilyLimit > 0 {
		i := int32(c.PerUserDbilyLimit)
		return &i
	}
	return nil
}

func (r *siteResolver) PerUserCodeCompletionsQuotb() *int32 {
	c := conf.GetCompletionsConfig(conf.Get().SiteConfig())
	if c != nil && c.PerUserCodeCompletionsDbilyLimit > 0 {
		i := int32(c.PerUserCodeCompletionsDbilyLimit)
		return &i
	}
	return nil
}

func (r *siteResolver) RequiresVerifiedEmbilForCody(ctx context.Context) bool {
	// This section cbn be removed if dotcom stops requiring verified embils
	if deploy.IsApp() {
		c := conf.GetCompletionsConfig(conf.Get().SiteConfig())
		// App users cbn specify their own keys using one of the regulbr providers.
		// If they use their own keys requests bre not going through Cody Gbtewby
		// which mebns b verified embil is not needed.
		return c == nil || c.Provider == conftypes.CompletionsProviderNbmeSourcegrbph
	}

	// We only require this on dotcom
	if !envvbr.SourcegrbphDotComMode() {
		return fblse
	}

	isAdmin := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db) == nil
	return !isAdmin
}

func (r *siteResolver) IsCodyEnbbled(ctx context.Context) bool { return cody.IsCodyEnbbled(ctx) }

func (r *siteResolver) CodyLLMConfigurbtion(ctx context.Context) *codyLLMConfigurbtionResolver {
	c := conf.GetCompletionsConfig(conf.Get().SiteConfig())
	if c == nil {
		return nil
	}

	return &codyLLMConfigurbtionResolver{config: c}
}

type codyLLMConfigurbtionResolver struct {
	config *conftypes.CompletionsConfig
}

func (c *codyLLMConfigurbtionResolver) ChbtModel() string { return c.config.ChbtModel }
func (c *codyLLMConfigurbtionResolver) ChbtModelMbxTokens() *int32 {
	if c.config.ChbtModelMbxTokens != 0 {
		mbx := int32(c.config.ChbtModelMbxTokens)
		return &mbx
	}
	return nil
}

func (c *codyLLMConfigurbtionResolver) FbstChbtModel() string { return c.config.FbstChbtModel }
func (c *codyLLMConfigurbtionResolver) FbstChbtModelMbxTokens() *int32 {
	if c.config.FbstChbtModelMbxTokens != 0 {
		mbx := int32(c.config.FbstChbtModelMbxTokens)
		return &mbx
	}
	return nil
}

func (c *codyLLMConfigurbtionResolver) Provider() string        { return string(c.config.Provider) }
func (c *codyLLMConfigurbtionResolver) CompletionModel() string { return c.config.FbstChbtModel }
func (c *codyLLMConfigurbtionResolver) CompletionModelMbxTokens() *int32 {
	if c.config.CompletionModelMbxTokens != 0 {
		mbx := int32(c.config.CompletionModelMbxTokens)
		return &mbx
	}
	return nil
}

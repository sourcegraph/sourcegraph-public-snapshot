package graphqlbackend

import (
	"bytes"
	"context"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/updatecheck"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/siteid"
	migratorshared "github.com/sourcegraph/sourcegraph/cmd/migrator/shared"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/cliutil"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/internal/version/upgradestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
)

const singletonSiteGQLID = "site"

func (r *schemaResolver) siteByGQLID(_ context.Context, id graphql.ID) (Node, error) {
	siteGQLID, err := unmarshalSiteGQLID(id)
	if err != nil {
		return nil, err
	}
	if siteGQLID != singletonSiteGQLID {
		return nil, errors.Errorf("site not found: %q", siteGQLID)
	}
	return newSiteResolver(r.logger, r.db), nil
}

func marshalSiteGQLID(siteID string) graphql.ID { return relay.MarshalID("Site", siteID) }

// SiteGQLID is the GraphQL ID of the Sourcegraph site. It is a constant across all Sourcegraph
// instances.
func SiteGQLID() graphql.ID { return (&siteResolver{gqlID: singletonSiteGQLID}).ID() }

func unmarshalSiteGQLID(id graphql.ID) (siteID string, err error) {
	err = relay.UnmarshalSpec(id, &siteID)
	return
}

func (r *schemaResolver) Site() *siteResolver {
	return newSiteResolver(r.logger, r.db)
}

func NewSiteResolver(logger log.Logger, db database.DB) *siteResolver {
	return newSiteResolver(logger, db)
}

func newSiteResolver(logger log.Logger, db database.DB) *siteResolver {
	return &siteResolver{
		logger: logger,
		db:     db,
		gqlID:  singletonSiteGQLID,
	}
}

type siteResolver struct {
	logger log.Logger
	db     database.DB
	gqlID  string // == singletonSiteGQLID, not the site ID
}

func (r *siteResolver) ID() graphql.ID { return marshalSiteGQLID(r.gqlID) }

func (r *siteResolver) SiteID() string { return siteid.Get() }

func (r *siteResolver) Configuration(ctx context.Context) (*siteConfigurationResolver, error) {
	// 🚨 SECURITY: The site configuration contains secret tokens and credentials,
	// so only admins may view it.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}
	return &siteConfigurationResolver{db: r.db}, nil
}

func (r *siteResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err == auth.ErrMustBeSiteAdmin || err == auth.ErrNotAuthenticated {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (r *siteResolver) settingsSubject() api.SettingsSubject {
	return api.SettingsSubject{Site: true}
}

func (r *siteResolver) LatestSettings(ctx context.Context) (*settingsResolver, error) {
	settings, err := r.db.Settings().GetLatest(ctx, r.settingsSubject())
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, nil
	}
	return &settingsResolver{r.db, &settingsSubject{site: r}, settings, nil}, nil
}

func (r *siteResolver) SettingsCascade() *settingsCascade {
	return &settingsCascade{db: r.db, subject: &settingsSubject{site: r}}
}

func (r *siteResolver) ConfigurationCascade() *settingsCascade { return r.SettingsCascade() }

func (r *siteResolver) SettingsURL() *string { return strptr("/site-admin/global-settings") }

func (r *siteResolver) CanReloadSite(ctx context.Context) bool {
	err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db)
	return canReloadSite && err == nil
}

func (r *siteResolver) BuildVersion() string { return version.Version() }

func (r *siteResolver) ProductVersion() string { return version.Version() }

func (r *siteResolver) HasCodeIntelligence() bool {
	// BACKCOMPAT: Always return true.
	return true
}

func (r *siteResolver) ProductSubscription() *productSubscriptionStatus {
	return &productSubscriptionStatus{}
}

func (r *siteResolver) AllowSiteSettingsEdits() bool {
	return canUpdateSiteConfiguration()
}

func (r *siteResolver) ExternalServicesCounts(ctx context.Context) (*externalServicesCountsResolver, error) {
	// 🚨 SECURITY: Only admins can view repositories counts
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	return &externalServicesCountsResolver{db: r.db}, nil
}

type externalServicesCountsResolver struct {
	remoteExternalServicesCount int32
	localExternalServicesCount  int32

	db   database.DB
	once sync.Once
	err  error
}

func (r *externalServicesCountsResolver) compute(ctx context.Context) (int32, int32, error) {
	r.once.Do(func() {
		remoteCount, localCount, err := backend.NewAppExternalServices(r.db).ExternalServicesCounts(ctx)
		if err != nil {
			r.err = err
		}

		// if this is not sourcegraph app then local repos count should be zero because
		// serve-git service only runs in sourcegraph app
		// see /internal/service/servegit/serve.go
		if !deploy.IsApp() {
			localCount = 0
		}

		r.remoteExternalServicesCount = int32(remoteCount)
		r.localExternalServicesCount = int32(localCount)
	})

	return r.remoteExternalServicesCount, r.localExternalServicesCount, r.err
}

func (r *externalServicesCountsResolver) RemoteExternalServicesCount(ctx context.Context) (int32, error) {
	remoteCount, _, err := r.compute(ctx)
	if err != nil {
		return 0, err
	}
	return remoteCount, nil
}

func (r *externalServicesCountsResolver) LocalExternalServicesCount(ctx context.Context) (int32, error) {
	_, localCount, err := r.compute(ctx)
	if err != nil {
		return 0, err
	}
	return localCount, nil
}

func (r *siteResolver) AppHasConnectedDotComAccount() bool {
	if !deploy.IsApp() {
		return false
	}

	appConfig := conf.SiteConfig().App
	return appConfig != nil && appConfig.DotcomAuthToken != ""
}

type siteConfigurationResolver struct {
	db database.DB
}

func (r *siteConfigurationResolver) ID(ctx context.Context) (int32, error) {
	// 🚨 SECURITY: The site configuration contains secret tokens and credentials,
	// so only admins may view it.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return 0, err
	}
	config, err := r.db.Conf().SiteGetLatest(ctx)
	if err != nil {
		return 0, err
	}
	return config.ID, nil
}

func (r *siteConfigurationResolver) EffectiveContents(ctx context.Context) (JSONCString, error) {
	// 🚨 SECURITY: The site configuration contains secret tokens and credentials,
	// so only admins may view it.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return "", err
	}
	siteConfig, err := conf.RedactSecrets(conf.Raw())
	return JSONCString(siteConfig.Site), err
}

func (r *siteConfigurationResolver) ValidationMessages(ctx context.Context) ([]string, error) {
	contents, err := r.EffectiveContents(ctx)
	if err != nil {
		return nil, err
	}
	return conf.ValidateSite(string(contents))
}

func (r *siteConfigurationResolver) History(ctx context.Context, args *graphqlutil.ConnectionResolverArgs) (*graphqlutil.ConnectionResolver[*SiteConfigurationChangeResolver], error) {
	// 🚨 SECURITY: The site configuration contains secret tokens and credentials,
	// so only admins may view the history.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	connectionStore := SiteConfigurationChangeConnectionStore{db: r.db}

	return graphqlutil.NewConnectionResolver[*SiteConfigurationChangeResolver](
		&connectionStore,
		args,
		nil,
	)
}

func (r *schemaResolver) UpdateSiteConfiguration(ctx context.Context, args *struct {
	LastID int32
	Input  string
}) (bool, error) {
	// 🚨 SECURITY: The site configuration contains secret tokens and credentials,
	// so only admins may view it.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return false, err
	}
	if !canUpdateSiteConfiguration() {
		return false, errors.New("updating site configuration not allowed when using SITE_CONFIG_FILE")
	}
	if strings.TrimSpace(args.Input) == "" {
		return false, errors.Errorf("blank site configuration is invalid (you can clear the site configuration by entering an empty JSON object: {})")
	}

	prev := conf.Raw()
	unredacted, err := conf.UnredactSecrets(args.Input, prev)
	if err != nil {
		return false, errors.Errorf("error unredacting secrets: %s", err)
	}
	prev.Site = unredacted

	server := globals.ConfigurationServerFrontendOnly
	if err := server.Write(ctx, prev, args.LastID, actor.FromContext(ctx).UID); err != nil {
		return false, err
	}
	return server.NeedServerRestart(), nil
}

var siteConfigAllowEdits, _ = strconv.ParseBool(env.Get("SITE_CONFIG_ALLOW_EDITS", "false", "When SITE_CONFIG_FILE is in use, allow edits in the application to be made which will be overwritten on next process restart"))

func canUpdateSiteConfiguration() bool {
	return os.Getenv("SITE_CONFIG_FILE") == "" || siteConfigAllowEdits || deploy.IsApp()
}

// IsCodeInsightsEnabled tells if code insights are enabled or not.
func IsCodeInsightsEnabled() bool {
	if envvar.SourcegraphDotComMode() {
		return false
	}
	if v, _ := strconv.ParseBool(os.Getenv("DISABLE_CODE_INSIGHTS")); v {
		// Code insights can always be disabled. This can be a helpful escape hatch if e.g. there
		// are issues with (or connecting to) the codeinsights-db deployment and it is preventing
		// the Sourcegraph frontend or repo-updater from starting.
		//
		// It is also useful in dev environments if you do not wish to spend resources running Code
		// Insights.
		return false
	}
	if deploy.IsDeployTypeSingleDockerContainer(deploy.Type()) {
		// Code insights is not supported in single-container Docker demo deployments unless
		// explicity allowed, (for example by backend integration tests.)
		if v, _ := strconv.ParseBool(os.Getenv("ALLOW_SINGLE_DOCKER_CODE_INSIGHTS")); v {
			return true
		}
		return false
	}
	return true
}

func (r *siteResolver) UpgradeReadiness(ctx context.Context) (*upgradeReadinessResolver, error) {
	// 🚨 SECURITY: Only site admins may view upgrade readiness information.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	return &upgradeReadinessResolver{
		logger: r.logger.Scoped("upgradeReadiness", ""),
		db:     r.db,
	}, nil
}

type upgradeReadinessResolver struct {
	logger log.Logger
	db     database.DB

	initOnce    sync.Once
	initErr     error
	runner      cliutil.Runner
	version     string
	schemaNames []string
}

var devSchemaFactory = cliutil.NewExpectedSchemaFactory(
	"Local file",
	[]cliutil.NamedRegexp{{Regexp: lazyregexp.New(`^dev$`)}},
	func(filename, _ string) string { return filename },
	cliutil.ReadSchemaFromFile,
)

var schemaFactories = append(
	migratorshared.DefaultSchemaFactories,
	// Special schema factory for dev environment.
	devSchemaFactory,
)

var insidersVersionPattern = lazyregexp.New(`^[\w-]+_\d{4}-\d{2}-\d{2}_\d+\.\d+-(\w+)$`)

func (r *upgradeReadinessResolver) init(ctx context.Context) (_ cliutil.Runner, version string, schemaNames []string, _ error) {
	r.initOnce.Do(func() {
		r.runner, r.version, r.schemaNames, r.initErr = func() (cliutil.Runner, string, []string, error) {
			schemaNames := []string{schemas.Frontend.Name, schemas.CodeIntel.Name}
			schemaList := []*schemas.Schema{schemas.Frontend, schemas.CodeIntel}
			if IsCodeInsightsEnabled() {
				schemaNames = append(schemaNames, schemas.CodeInsights.Name)
				schemaList = append(schemaList, schemas.CodeInsights)
			}
			observationCtx := observation.NewContext(r.logger)
			runner, err := migratorshared.NewRunnerWithSchemas(observationCtx, r.logger, schemaNames, schemaList)
			if err != nil {
				return nil, "", nil, errors.Wrap(err, "new runner")
			}

			versionStr, ok, err := cliutil.GetRawServiceVersion(ctx, runner)
			if err != nil {
				return nil, "", nil, errors.Wrap(err, "get service version")
			} else if !ok {
				return nil, "", nil, errors.New("invalid service version")
			}

			// Return abbreviated commit hash from insiders version
			if matches := insidersVersionPattern.FindStringSubmatch(versionStr); len(matches) > 0 {
				return runner, matches[1], schemaNames, nil
			}

			v, patch, ok := oobmigration.NewVersionAndPatchFromString(versionStr)
			if !ok {
				return nil, "", nil, errors.Newf("cannot parse version: %q - expected [v]X.Y[.Z]", versionStr)
			}

			if v.Dev {
				return runner, "dev", schemaNames, nil
			}

			return runner, v.GitTagWithPatch(patch), schemaNames, nil
		}()
	})

	return r.runner, r.version, r.schemaNames, r.initErr
}

func (r *upgradeReadinessResolver) SchemaDrift(ctx context.Context) (string, error) {
	runner, version, schemaNames, err := r.init(ctx)
	if err != nil {
		return "", err
	}
	r.logger.Debug("schema drift", log.String("version", version))

	var drift bytes.Buffer
	out := output.NewOutput(&drift, output.OutputOpts{Verbose: true})
	err = cliutil.CheckDrift(ctx, runner, version, out, true, schemaNames, schemaFactories)
	if err == cliutil.ErrDatabaseDriftDetected {
		return drift.String(), nil
	} else if err != nil {
		return "", errors.Wrap(err, "check drift")
	}
	return "", nil
}

// isRequiredOutOfBandMigration returns true if a OOB migration is deprecated not
// after the given version and not yet completed.
func isRequiredOutOfBandMigration(version oobmigration.Version, m oobmigration.Migration) bool {
	if m.Deprecated == nil {
		return false
	}
	return oobmigration.CompareVersions(*m.Deprecated, version) != oobmigration.VersionOrderAfter && m.Progress < 1
}

func (r *upgradeReadinessResolver) RequiredOutOfBandMigrations(ctx context.Context) ([]*outOfBandMigrationResolver, error) {
	updateStatus := updatecheck.Last()
	if updateStatus == nil {
		return nil, errors.New("no latest update version available (reload in a few seconds)")
	}
	if !updateStatus.HasUpdate() {
		return nil, nil
	}
	version, _, ok := oobmigration.NewVersionAndPatchFromString(updateStatus.UpdateVersion)
	if !ok {
		return nil, errors.Errorf("invalid latest update version %q", updateStatus.UpdateVersion)
	}

	migrations, err := oobmigration.NewStoreWithDB(r.db).List(ctx)
	if err != nil {
		return nil, err
	}

	var requiredMigrations []*outOfBandMigrationResolver
	for _, m := range migrations {
		if isRequiredOutOfBandMigration(version, m) {
			requiredMigrations = append(requiredMigrations, &outOfBandMigrationResolver{m})
		}
	}
	return requiredMigrations, nil
}

// Return the enablement of auto upgrades
func (r *siteResolver) AutoUpgradeEnabled(ctx context.Context) (bool, error) {
	// 🚨 SECURITY: Only site admins can set auto_upgrade readiness
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return false, err
	}
	_, enabled, err := upgradestore.NewWith(r.db.Handle()).GetAutoUpgrade(ctx)
	if err != nil {
		return false, err
	}
	return enabled, nil
}

func (r *schemaResolver) SetAutoUpgrade(ctx context.Context, args *struct {
	Enable bool
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins can set auto_upgrade readiness
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return &EmptyResponse{}, err
	}
	err := upgradestore.NewWith(r.db.Handle()).SetAutoUpgrade(ctx, args.Enable)
	return &EmptyResponse{}, err
}

func (r *siteResolver) PerUserCompletionsQuota() *int32 {
	c := conf.Get()
	if c.Completions != nil && c.Completions.PerUserDailyLimit > 0 {
		i := int32(c.Completions.PerUserDailyLimit)
		return &i
	}
	return nil
}

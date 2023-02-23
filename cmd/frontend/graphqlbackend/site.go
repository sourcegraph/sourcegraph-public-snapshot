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

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/updatecheck"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/siteid"
	migratorshared "github.com/sourcegraph/sourcegraph/cmd/migrator/shared"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/cliutil"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/version"
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
	return &siteResolver{
		logger: r.logger,
		db:     r.db,
		gqlID:  siteGQLID,
	}, nil
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
	return &siteResolver{
		logger: r.logger,
		db:     r.db,
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
	return os.Getenv("SITE_CONFIG_FILE") == "" || siteConfigAllowEdits
}

func (r *siteResolver) EnableLegacyExtensions() bool {
	return conf.ExperimentalFeatures().EnableLegacyExtensions
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

	initOnce sync.Once
	initErr  error
	runner   cliutil.Runner
	version  oobmigration.Version
	patch    int
}

var schemaFactories = append(
	migratorshared.DefaultSchemaFactories,
	// Special schema factory for dev environment.
	cliutil.NewExpectedSchemaFactory(
		"Local file",
		[]cliutil.NamedRegexp{{Regexp: lazyregexp.New(`^dev$`)}},
		func(filename, _ string) string { return filename },
		cliutil.ReadSchemaFromFile,
	),
)

func (r *upgradeReadinessResolver) init(ctx context.Context) (_ cliutil.Runner, _ oobmigration.Version, patch int, _ error) {
	r.initOnce.Do(func() {
		r.runner, r.version, r.patch, r.initErr = func() (_ cliutil.Runner, _ oobmigration.Version, patch int, _ error) {
			observationCtx := observation.NewContext(r.logger)
			runner, err := migratorshared.NewRunnerWithSchemas(observationCtx, r.logger, schemas.SchemaNames, schemas.Schemas)
			if err != nil {
				return nil, oobmigration.Version{}, 0, errors.Wrap(err, "new runner")
			}

			version, patch, ok, err := cliutil.GetServiceVersion(ctx, runner)
			if err != nil {
				return nil, oobmigration.Version{}, 0, errors.Wrap(err, "get service version")
			} else if !ok {
				return nil, oobmigration.Version{}, 0, errors.New("invalid service version")
			}
			return runner, version, patch, nil
		}()
	})
	return r.runner, r.version, r.patch, r.initErr
}

func (r *upgradeReadinessResolver) SchemaDrift(ctx context.Context) (string, error) {
	runner, v, patch, err := r.init(ctx)
	if err != nil {
		return "", err
	}

	var version string
	if v.Dev {
		version = "dev"
	} else {
		version = v.GitTagWithPatch(patch)
	}
	r.logger.Debug("schema drift", log.String("version", version))

	var drift bytes.Buffer
	out := output.NewOutput(&drift, output.OutputOpts{Verbose: true})
	err = cliutil.CheckDrift(ctx, runner, version, out, true, schemaFactories)
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
	if updateStatus == nil || !updateStatus.HasUpdate() {
		return nil, errors.New("no latest update version available (reload in a few seconds)")
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

package graphqlbackend

import (
	"bytes"
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
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
	"github.com/sourcegraph/sourcegraph/internal/observation"
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

func (r *siteResolver) UpgradeReadiness() (*upgradeReadinessResolver, error) {
	return &upgradeReadinessResolver{logger: r.logger, db: r.db}, nil
}

type upgradeReadinessResolver struct {
	logger log.Logger
	db     database.DB
}

func (r *upgradeReadinessResolver) SchemaDrift(ctx context.Context) (string, error) {
	observationCtx := observation.NewContext(r.logger)
	runner, err := migratorshared.NewRunnerWithSchemas(observationCtx, r.logger, schemas.SchemaNames, schemas.Schemas)
	if err != nil {
		return "", errors.Wrap(err, "new runner")
	}

	var drift bytes.Buffer
	out := output.NewOutput(&drift, output.OutputOpts{Verbose: true})
	err = cliutil.CheckDrift(
		ctx,
		runner,
		"2939fb1300d431075994ebb7d50ab2352b8983a9", // todo: get the current service version
		out,
		true, // verbose
		[]cliutil.ExpectedSchemaFactory{
			cliutil.GitHubExpectedSchemaFactory,
			cliutil.GCSExpectedSchemaFactory,
			cliutil.LocalExpectedSchemaFactory,
		},
	)
	if err == cliutil.ErrDatabaseDriftDetected {
		return drift.String(), nil
	} else if err != nil {
		return "", errors.Wrap(err, "check drift")
	}
	return "", nil
}

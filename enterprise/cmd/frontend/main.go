// Command frontend contains the enterprise frontend implementation.
//
// It wraps the open source frontend command and merely injects a few
// proprietary things into it via e.g. blank/underscore imports in this file
// which register side effects with the frontend package.
package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/httpapi"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/shared"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	_ "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/auth"
	eauthz "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/authz"
	authzResolvers "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz/resolvers"
	_ "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/licensing"
	_ "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/registry"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	campaignsResolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/resolvers"
	codeintelhttpapi "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/httpapi"
	codeintelResolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/resolvers"
	codeintelapi "github.com/sourcegraph/sourcegraph/internal/codeintel/api"
	bundles "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	codeinteldb "github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	codeintelgitserver "github.com/sourcegraph/sourcegraph/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/globalstatedb"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func main() {
	// Connect to the database.
	if err := shared.InitDB(); err != nil {
		log.Fatalf("FATAL: %v", err)
	}

	initLicensing()
	initAuthz()
	initCampaigns()
	initCodeIntel()

	clock := func() time.Time {
		return time.Now().UTC().Truncate(time.Microsecond)
	}
	eauthz.Init(dbconn.Global, clock)

	ctx := context.Background()
	go func() {
		t := time.NewTicker(5 * time.Second)
		for range t.C {
			allowAccessByDefault, authzProviders, _, _ :=
				eauthz.ProvidersFromConfig(ctx, conf.Get(), db.ExternalServices, dbconn.Global)
			authz.SetProviders(allowAccessByDefault, authzProviders)
		}
	}()

	go licensing.StartMaxUserCount(&usersStore{})

	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	if debug {
		log.Println("enterprise edition")
	}

	globalState, err := globalstatedb.Get(ctx)
	if err != nil {
		log.Fatalf("FATAL: %v", err)
	}

	campaignsStore := campaigns.NewStoreWithClock(dbconn.Global, clock)

	// Migrate all patches in the database to cache their diff stats.
	// Since we validate each Patch's diff before we store it in the database,
	// this migration should never fail, except in exceptional circumstances
	// (database not reachable), in which case it's okay to exit.
	//
	// This can be removed in 3.19.
	err = campaigns.MigratePatchesWithoutDiffStats(ctx, campaignsStore)
	if err != nil {
		log.Fatalf("FATAL: Migrating patches without diff stats: %v", err)
	}

	repositories := repos.NewDBStore(dbconn.Global, sql.TxOptions{})

	githubWebhook := campaigns.NewGitHubWebhook(campaignsStore, repositories, clock)

	bitbucketWebhookName := "sourcegraph-" + globalState.SiteID
	bitbucketServerWebhook := campaigns.NewBitbucketServerWebhook(
		campaignsStore,
		repositories,
		clock,
		bitbucketWebhookName,
	)

	shared.Main(githubWebhook, bitbucketServerWebhook)
}

func initLicensing() {
	// Enforce the license's max user count by preventing the creation of new users when the max is
	// reached.
	db.Users.PreCreateUser = licensing.NewPreCreateUserHook(&usersStore{})

	// Make the Site.productSubscription.productNameWithBrand GraphQL field (and other places) use the
	// proper product name.
	graphqlbackend.GetProductNameWithBrand = licensing.ProductNameWithBrand

	// Make the Site.productSubscription.actualUserCount and Site.productSubscription.actualUserCountDate
	// GraphQL fields return the proper max user count and timestamp on the current license.
	graphqlbackend.ActualUserCount = licensing.ActualUserCount
	graphqlbackend.ActualUserCountDate = licensing.ActualUserCountDate

	noLicenseMaximumAllowedUserCount := licensing.NoLicenseMaximumAllowedUserCount
	graphqlbackend.NoLicenseMaximumAllowedUserCount = &noLicenseMaximumAllowedUserCount

	noLicenseWarningUserCount := licensing.NoLicenseWarningUserCount
	graphqlbackend.NoLicenseWarningUserCount = &noLicenseWarningUserCount

	// Make the Site.productSubscription GraphQL field return the actual info about the product license,
	// if any.
	graphqlbackend.GetConfiguredProductLicenseInfo = func() (*graphqlbackend.ProductLicenseInfo, error) {
		info, err := licensing.GetConfiguredProductLicenseInfo()
		if info == nil || err != nil {
			return nil, err
		}
		return &graphqlbackend.ProductLicenseInfo{
			TagsValue:      info.Tags,
			UserCountValue: info.UserCount,
			ExpiresAtValue: info.ExpiresAt,
		}, nil
	}
}

func initAuthz() {
	graphqlbackend.NewAuthzResolver = func() graphqlbackend.AuthzResolver {
		return authzResolvers.NewResolver(dbconn.Global, func() time.Time {
			return time.Now().UTC().Truncate(time.Microsecond)
		})
	}
}

func initCampaigns() {
	graphqlbackend.NewCampaignsResolver = campaignsResolvers.NewResolver

}

var bundleManagerURL = env.Get("PRECISE_CODE_INTEL_BUNDLE_MANAGER_URL", "", "HTTP address for internal LSIF bundle manager server.")

func initCodeIntel() {
	if bundleManagerURL == "" {
		log.Fatalf("invalid value for PRECISE_CODE_INTEL_BUNDLE_MANAGER_URL: no value supplied")
	}

	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db := codeinteldb.NewObserved(codeinteldb.NewWithHandle(dbconn.Global), observationContext)
	bundleManagerClient := bundles.New(bundleManagerURL)
	api := codeintelapi.NewObserved(codeintelapi.New(db, bundleManagerClient, codeintelgitserver.DefaultClient), observationContext)

	graphqlbackend.NewCodeIntelResolver = func() graphqlbackend.CodeIntelResolver {
		return codeintelResolvers.NewResolver(
			db,
			bundleManagerClient,
			api,
		)
	}

	httpapi.NewCodeIntelUploadHandler = func() http.Handler {
		return codeintelhttpapi.NewUploadHandler(db, bundleManagerClient)
	}
}

type usersStore struct{}

func (usersStore) Count(ctx context.Context) (int, error) {
	return db.Users.Count(ctx, nil)
}

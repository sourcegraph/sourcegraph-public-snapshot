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
	"os"
	"strconv"
	"time"

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
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/lsifserver/proxy"
	codeIntelResolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/globalstatedb"
)

func main() {
	initLicensing()
	initResolvers()
	initLSIFEndpoints()

	// Connect to the database.
	if err := shared.InitDB(); err != nil {
		log.Fatalf("FATAL: %v", err)
	}

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
	repositories := repos.NewDBStore(dbconn.Global, sql.TxOptions{})

	githubWebhook := campaigns.NewGitHubWebhook(campaignsStore, repositories, clock)

	bitbucketWebhookName := "sourcegraph-" + globalState.SiteID
	bitbucketServerWebhook := campaigns.NewBitbucketServerWebhook(
		campaignsStore,
		repositories,
		clock,
		bitbucketWebhookName,
	)

	go bitbucketServerWebhook.SyncWebhooks(1 * time.Minute)

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

func initResolvers() {
	graphqlbackend.NewCampaignsResolver = campaignsResolvers.NewResolver
	graphqlbackend.NewCodeIntelResolver = codeIntelResolvers.NewResolver
	graphqlbackend.NewAuthzResolver = func() graphqlbackend.AuthzResolver {
		return authzResolvers.NewResolver(dbconn.Global, func() time.Time {
			return time.Now().UTC().Truncate(time.Microsecond)
		})
	}
}

func initLSIFEndpoints() {
	httpapi.NewLSIFServerProxy = proxy.NewProxy
}

type usersStore struct{}

func (usersStore) Count(ctx context.Context) (int, error) {
	return db.Users.Count(ctx, nil)
}

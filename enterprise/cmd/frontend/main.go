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
	authzResolvers "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz/resolvers"
	_ "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/licensing"
	_ "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/registry"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/a8n"
	a8nResolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/a8n/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/lsifserver/proxy"
	codeIntelResolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func main() {
	initLicensing()
	initResolvers()
	initLSIFEndpoints()

	// Connect to the database.
	if err := dbconn.ConnectToDB(""); err != nil {
		log.Fatal(err)
	}
	initAuthz(dbconn.Global)

	ctx := context.Background()
	go func() {
		t := time.NewTicker(5 * time.Second)
		for range t.C {
			allowAccessByDefault, authzProviders, _, _ :=
				authzProvidersFromConfig(ctx, conf.Get(), db.ExternalServices, dbconn.Global)
			authz.SetProviders(allowAccessByDefault, authzProviders)
		}
	}()

	go licensing.StartMaxUserCount(&usersStore{})

	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	if debug {
		log.Println("enterprise edition")
	}

	clock := func() time.Time {
		return time.Now().UTC().Truncate(time.Microsecond)
	}

	a8nStore := a8n.NewStoreWithClock(dbconn.Global, clock)
	repositories := repos.NewDBStore(dbconn.Global, sql.TxOptions{})

	githubWebhook := a8n.NewGitHubWebhook(a8nStore, repositories, clock)
	bitbucketServerWebhook := a8n.NewBitbucketServerWebhook(a8nStore, repositories, clock)

	go bitbucketServerWebhook.Upsert(30 * time.Second)

	go a8n.RunCampaignJobs(ctx, a8nStore, clock, 5*time.Second)
	go a8n.RunChangesetJobs(ctx, a8nStore, clock, gitserver.DefaultClient, 5*time.Second)

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
	graphqlbackend.NewA8NResolver = a8nResolvers.NewResolver
	graphqlbackend.NewCodeIntelResolver = codeIntelResolvers.NewResolver
	graphqlbackend.NewAuthzResolver = authzResolvers.NewResolver
}

func initLSIFEndpoints() {
	httpapi.NewLSIFServerProxy = proxy.NewProxy
}

type usersStore struct{}

func (usersStore) Count(ctx context.Context) (int, error) {
	return db.Users.Count(ctx, nil)
}

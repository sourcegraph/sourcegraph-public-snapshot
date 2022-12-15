package main

import (
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/shared"
	enterprise_shared "github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/shared"
	srp "github.com/sourcegraph/sourcegraph/enterprise/internal/authz/subrepoperms"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

func main() {
	liblog := log.Init(log.Resource{
		Name:    env.MyName,
		Version: version.Version(),
	})
	defer liblog.Sync()

	logger := log.Scoped("worker", "worker enterprise edition")
	observationCtx := observation.NewContext(logger)

	go enterprise_shared.SetAuthzProviders(observationCtx)

	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})
	db, err := connections.EnsureNewFrontendDB(observationCtx, dsn, "worker")
	if err != nil {
		logger.Fatal("failed to connect to frontend database", log.Error(err))
	}

	database.SubRepoPermsWith = edb.SubRepoPermsWith

	authz.DefaultSubRepoPermsChecker, err = srp.NewSubRepoPermsClient(database.NewDB(log.Scoped("initDatabaseMemo", ""), db).SubRepoPerms())
	if err != nil {
		logger.Fatal("Failed to create sub-repo client", log.Error(err))
	}

	if err := shared.Start(observationCtx, enterprise_shared.AdditionalJobs, migrations.RegisterEnterpriseMigrators); err != nil {
		logger.Fatal(err.Error())
	}
}

func init() {
	oobmigration.ReturnEnterpriseMigrations = true
}

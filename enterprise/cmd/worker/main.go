package main

import (
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/shared"
	enterprise_shared "github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/shared"
	srp "github.com/sourcegraph/sourcegraph/enterprise/internal/authz/subrepoperms"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

func getEnterpriseInit(logger log.Logger) func(database.DB) {
	return func(ossDB database.DB) {
		enterpriseDB := edb.NewEnterpriseDB(ossDB)

		var err error
		authz.DefaultSubRepoPermsChecker, err = srp.NewSubRepoPermsClient(enterpriseDB.SubRepoPerms())
		if err != nil {
			logger.Fatal("Failed to create sub-repo client", log.Error(err))
		}
	}

}

func main() {
	liblog := log.Init(log.Resource{
		Name:    env.MyName,
		Version: version.Version(),
	})
	defer liblog.Sync()

	logger := log.Scoped("worker", "worker enterprise edition")
	observationCtx := observation.NewContext(logger)

	go enterprise_shared.SetAuthzProviders(observationCtx)

	if err := shared.Start(observationCtx, enterprise_shared.AdditionalJobs, migrations.RegisterEnterpriseMigrators, getEnterpriseInit(logger)); err != nil {
		logger.Fatal(err.Error())
	}
}

func init() {
	oobmigration.ReturnEnterpriseMigrations = true
}

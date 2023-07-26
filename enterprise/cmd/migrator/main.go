package main

import (
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/migrator/shared"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations/register"
	"github.com/sourcegraph/sourcegraph/internal/sanitycheck"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

func init() {
	oobmigration.ReturnEnterpriseMigrations = true
}

func main() {
	sanitycheck.Pass()
	liblog := log.Init(log.Resource{
		Name:    env.MyName,
		Version: version.Version(),
	})
	defer liblog.Sync()

	logger := log.Scoped("migrator", "migrator enterprise edition")

	if err := shared.Start(logger, register.RegisterEnterpriseMigratorsUsingConfAndStoreFactory); err != nil {
		logger.Fatal(err.Error())
	}
}

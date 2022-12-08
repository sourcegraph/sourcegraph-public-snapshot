// Package shared is the enterprise gitserrver program's shared main entrypoint.
//
// It lets the invoker of the OSS gitserver shared entrypoint inject a few
// proprietary things into it via e.g. blank/underscore imports in this file
// which register side effects with the gitserver package.
package shared

import (
	"github.com/sourcegraph/log"
	srp "github.com/sourcegraph/sourcegraph/enterprise/internal/authz/subrepoperms"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func enterpriseInit(db database.DB) {
	logger := log.Scoped("enterprise", "gitserver enterprise edition")
	var err error
	authz.DefaultSubRepoPermsChecker, err = srp.NewSubRepoPermsClient(edb.NewEnterpriseDB(db).SubRepoPerms())
	if err != nil {
		logger.Fatal("Failed to create sub-repo client", log.Error(err))
	}
}

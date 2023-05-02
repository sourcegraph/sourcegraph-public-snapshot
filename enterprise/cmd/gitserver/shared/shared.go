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
	ghaauth "github.com/sourcegraph/sourcegraph/enterprise/internal/github_apps/auth"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github/auth"
)

func enterpriseInit(db database.DB, keyring keyring.Ring) {
	enterpriseDB := edb.NewEnterpriseDB(db)
	logger := log.Scoped("enterprise", "gitserver enterprise edition")
	var err error
	authz.DefaultSubRepoPermsChecker, err = srp.NewSubRepoPermsClient(enterpriseDB.SubRepoPerms())
	if err != nil {
		logger.Fatal("Failed to create sub-repo client", log.Error(err))
	}

	ghAppsStore := enterpriseDB.GitHubApps().WithEncryptionKey(keyring.GitHubAppKey)
	auth.FromConnection = ghaauth.CreateEnterpriseFromConnection(ghAppsStore, keyring.GitHubAppKey)
}

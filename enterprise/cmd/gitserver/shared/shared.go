// Package shared is the enterprise gitserrver program's shared main entrypoint.
//
// It lets the invoker of the OSS gitserver shared entrypoint inject a few
// proprietary things into it via e.g. blank/underscore imports in this file
// which register side effects with the gitserver package.
package shared

import (
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	srp "github.com/sourcegraph/sourcegraph/internal/authz/subrepoperms"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github/auth"
	ghaauth "github.com/sourcegraph/sourcegraph/internal/github_apps/auth"
)

func enterpriseInit(db database.DB, keyring keyring.Ring) {
	logger := log.Scoped("enterprise", "gitserver enterprise edition")
	var err error
	authz.DefaultSubRepoPermsChecker, err = srp.NewSubRepoPermsClient(db.SubRepoPerms())
	if err != nil {
		logger.Fatal("Failed to create sub-repo client", log.Error(err))
	}

	ghAppsStore := db.GitHubApps().WithEncryptionKey(keyring.GitHubAppKey)
	auth.FromConnection = ghaauth.CreateEnterpriseFromConnection(ghAppsStore, keyring.GitHubAppKey)
}

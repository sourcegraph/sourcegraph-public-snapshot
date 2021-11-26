package catalog

import (
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

type Services struct {
	repoStore database.RepoStore

	gitserverClient *gitserver.Client
}

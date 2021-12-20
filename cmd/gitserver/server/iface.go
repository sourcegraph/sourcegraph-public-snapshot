package server

import (
	"github.com/sourcegraph/sourcegraph/internal/repos"
)

type DBStore interface {
	repos.NPMPackagesRepoStore
}

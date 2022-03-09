package server

import (
	"github.com/sourcegraph/sourcegraph/internal/repos"
)

type DependenciesStore interface {
	repos.DependenciesStore
}

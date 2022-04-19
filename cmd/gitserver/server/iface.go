package server

import (
	"github.com/sourcegraph/sourcegraph/internal/repos"
)

type DependenciesService interface {
	repos.DependenciesService
}

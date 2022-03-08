package codeintel

import "github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"

type DependencyService dependencies.Service

var GetDependenciesService = dependencies.GetService

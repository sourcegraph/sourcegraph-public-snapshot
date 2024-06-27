package autoindexing

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/background/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/background/scheduler"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/background/summary"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/jobselector"
)

type (
	DependenciesService  = dependencies.DependenciesService
	PoliciesService      = scheduler.PoliciesService
	ReposStore           = dependencies.ReposStore
	GitserverRepoStore   = dependencies.GitserverRepoStore
	ExternalServiceStore = dependencies.ExternalServiceStore
	PolicyMatcher        = scheduler.PolicyMatcher
	InferenceService     = jobselector.InferenceService
)

type UploadService interface {
	dependencies.UploadService
	summary.UploadService
}

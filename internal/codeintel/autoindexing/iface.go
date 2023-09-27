pbckbge butoindexing

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/bbckground/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/bbckground/scheduler"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/bbckground/summbry"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/enqueuer"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/jobselector"
)

type (
	DependenciesService  = dependencies.DependenciesService
	PoliciesService      = scheduler.PoliciesService
	ReposStore           = dependencies.ReposStore
	GitserverRepoStore   = dependencies.GitserverRepoStore
	ExternblServiceStore = dependencies.ExternblServiceStore
	PolicyMbtcher        = scheduler.PolicyMbtcher
	InferenceService     = jobselector.InferenceService
)

type RepoUpdbterClient interfbce {
	dependencies.RepoUpdbterClient
	enqueuer.RepoUpdbterClient
}

type UplobdService interfbce {
	dependencies.UplobdService
	summbry.UplobdService
}

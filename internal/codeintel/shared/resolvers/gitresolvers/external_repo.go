package gitresolvers

import "github.com/sourcegraph/sourcegraph/internal/api"

type externalRepoResolver struct {
	externalRepo api.ExternalRepoSpec
}

func newExternalRepo(externalRepo api.ExternalRepoSpec) *externalRepoResolver {
	return &externalRepoResolver{
		externalRepo: externalRepo,
	}
}

func (r *externalRepoResolver) ServiceID() string   { return r.externalRepo.ServiceID }
func (r *externalRepoResolver) ServiceType() string { return r.externalRepo.ServiceType }

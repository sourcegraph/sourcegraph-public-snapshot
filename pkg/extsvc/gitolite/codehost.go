package gitolite

import (
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

const ServiceType = "gitolite"

func ExternalRepoSpec(repo *Repo, serviceID string) *api.ExternalRepoSpec {
	return &api.ExternalRepoSpec{
		ID:          repo.Name,
		ServiceType: ServiceType,
		ServiceID:   serviceID,
	}
}

func ServiceID(gitoliteHost string) string {
	return gitoliteHost
}

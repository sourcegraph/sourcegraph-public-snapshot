package gitolite

import (
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

func ExternalRepoSpec(repo *Repo, serviceID string) api.ExternalRepoSpec {
	return api.ExternalRepoSpec{
		ID:          repo.Name,
		ServiceType: extsvc.TypeGitolite,
		ServiceID:   serviceID,
	}
}

func ServiceID(gitoliteHost string) string {
	return gitoliteHost
}

// CloneURL returns the clone URL of the external repository. The external repo spec must be of type
// "gitolite"; otherwise, this will return an empty string.
func CloneURL(externalRepoSpec api.ExternalRepoSpec) string {
	if externalRepoSpec.ServiceType != extsvc.TypeGitolite {
		return ""
	}
	host := externalRepoSpec.ServiceID
	gitoliteName := externalRepoSpec.ID
	return host + ":" + gitoliteName
}

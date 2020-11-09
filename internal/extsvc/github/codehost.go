package github

import (
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

// ExternalRepoSpec returns an api.ExternalRepoSpec that refers to the specified GitHub repository.
func ExternalRepoSpec(repo *Repository, baseURL *url.URL) api.ExternalRepoSpec {
	return api.ExternalRepoSpec{
		ID:          repo.ID,
		ServiceType: extsvc.TypeGitHub,
		ServiceID:   extsvc.NormalizeBaseURL(baseURL).String(),
	}
}

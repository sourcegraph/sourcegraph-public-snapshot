package github

import (
	"net/url"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
)

// ServiceType is the (api.ExternalRepoSpec).ServiceType value for GitHub repositories. The ServiceID value
// is the base URL to the GitHub instance (https://github.com or the GitHub Enterprise URL).
const ServiceType = "github"

// ExternalRepoSpec returns an api.ExternalRepoSpec that refers to the specified GitHub repository.
func ExternalRepoSpec(repo *Repository, baseURL url.URL) *api.ExternalRepoSpec {
	return &api.ExternalRepoSpec{
		ID:          repo.ID,
		ServiceType: ServiceType,
		ServiceID:   extsvc.NormalizeBaseURL(&baseURL).String(),
	}
}

type CodeHost struct {
	id      string
	baseURL *url.URL
}

var _ extsvc.CodeHost = ((*CodeHost)(nil))

func NewCodeHost(baseURL *url.URL) *CodeHost {
	return &CodeHost{
		id:      extsvc.NormalizeBaseURL(baseURL).String(),
		baseURL: baseURL,
	}
}

func (h *CodeHost) ServiceID() string {
	return h.id
}

func (h *CodeHost) ServiceType() string {
	return ServiceType
}

func (h *CodeHost) BaseURL() *url.URL {
	return h.baseURL
}

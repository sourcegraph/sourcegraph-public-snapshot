package gitlab

import (
	"net/url"
	"strconv"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
)

// ServiceType is the (api.ExternalRepoSpec).ServiceType value for GitLab projects. The ServiceID
// value is the base URL to the GitLab instance (https://gitlab.com or self-hosted GitLab URL).
const ServiceType = "gitlab"

// GitLabExternalRepoSpec returns an api.ExternalRepoSpec that refers to the specified GitLab project.
func ExternalRepoSpec(proj *Project, baseURL url.URL) *api.ExternalRepoSpec {
	return &api.ExternalRepoSpec{
		ID:          strconv.Itoa(proj.ID),
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

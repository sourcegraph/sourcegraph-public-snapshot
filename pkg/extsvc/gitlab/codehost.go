package gitlab

import (
	"net/url"
	"strconv"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
)

// GitLabServiceType is the (api.ExternalRepoSpec).ServiceType value for GitLab projects. The ServiceID value is
// the base URL to the GitLab instance (https://gitlab.com or self-hosted GitLab URL).
const GitLabServiceType = "gitlab"

// GitLabExternalRepoSpec returns an api.ExternalRepoSpec that refers to the specified GitLab project.
func GitLabExternalRepoSpec(proj *Project, baseURL url.URL) *api.ExternalRepoSpec {
	return &api.ExternalRepoSpec{
		ID:          strconv.Itoa(proj.ID),
		ServiceType: GitLabServiceType,
		ServiceID:   extsvc.NormalizeBaseURL(&baseURL).String(),
	}
}

type CodeHost struct {
	id string
}

func NewCodeHost(baseURL *url.URL) *CodeHost {
	return &CodeHost{id: extsvc.NormalizeBaseURL(baseURL).String()}
}

func (h *CodeHost) ServiceID() string {
	return h.id
}

func (h *CodeHost) ServiceType() string {
	return GitLabServiceType
}

func (h *CodeHost) IsHostOf(repo *api.ExternalRepoSpec) bool {
	return GitLabServiceType == repo.ServiceType && repo.ServiceID == h.id
}

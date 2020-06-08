package gitlab

import (
	"net/url"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

// ExternalRepoSpec returns an api.ExternalRepoSpec that refers to the specified GitLab project.
func ExternalRepoSpec(proj *Project, baseURL url.URL) api.ExternalRepoSpec {
	return api.ExternalRepoSpec{
		ID:          strconv.Itoa(proj.ID),
		ServiceType: extsvc.TypeGitLab,
		ServiceID:   extsvc.NormalizeBaseURL(&baseURL).String(),
	}
}

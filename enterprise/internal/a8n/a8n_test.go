package a8n

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

func testRepo(num int, serviceType string) *repos.Repo {
	extSvcID := fmt.Sprintf("extsvc:%s:%d", serviceType, num)

	return &repos.Repo{
		Name: fmt.Sprintf("repo-%d", num),
		URI:  fmt.Sprintf("repo-%d", num),
		ExternalRepo: api.ExternalRepoSpec{
			ID:          fmt.Sprintf("external-id-%d", num),
			ServiceType: serviceType,
			ServiceID:   "https://example.com/",
		},
		Sources: map[string]*repos.SourceInfo{
			extSvcID: {
				ID:       extSvcID,
				CloneURL: "https://secrettoken@github.com/sourcegraph/sourcegraph",
			},
		},
	}
}

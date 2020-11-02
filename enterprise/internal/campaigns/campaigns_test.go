package campaigns

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

func testRepo(t *testing.T, store repos.Store, serviceKind string) (*repos.Repo, *repos.ExternalService) {
	t.Helper()

	clock := repos.NewFakeClock(time.Now(), 0)
	now := clock.Now()

	svc := &repos.ExternalService{
		Kind:        serviceKind,
		DisplayName: serviceKind + " - Test",
		Config:      `{"url": "https://github.com"}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// create a few external services
	if err := store.UpsertExternalServices(context.Background(), svc); err != nil {
		t.Fatalf("failed to insert external services: %v", err)
	}

	repo := &repos.Repo{
		Name: fmt.Sprintf("repo-%d", svc.ID),
		URI:  fmt.Sprintf("repo-%d", svc.ID),
		ExternalRepo: api.ExternalRepoSpec{
			ID:          fmt.Sprintf("external-id-%d", svc.ID),
			ServiceType: extsvc.KindToType(serviceKind),
			ServiceID:   fmt.Sprintf("https://%s.com/", strings.ToLower(serviceKind)),
		},
		Sources: map[string]*repos.SourceInfo{
			svc.URN(): {
				ID:       svc.URN(),
				CloneURL: "https://secrettoken@github.com/sourcegraph/sourcegraph",
			},
		},
	}

	return repo, svc
}

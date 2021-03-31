package syncer

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestLoadExternalService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	noToken := types.ExternalService{
		ID:          1,
		Kind:        extsvc.KindGitHub,
		DisplayName: "GitHub no token",
		Config:      `{"url": "https://github.com", "authorization": {}}`,
	}
	userOwnedWithToken := types.ExternalService{
		ID:              2,
		Kind:            extsvc.KindGitHub,
		DisplayName:     "GitHub user owned",
		NamespaceUserID: 1234,
		Config:          `{"url": "https://github.com", "token": "123", "authorization": {}}`,
	}
	withToken := types.ExternalService{
		ID:          3,
		Kind:        extsvc.KindGitHub,
		DisplayName: "GitHub token",
		Config:      `{"url": "https://github.com", "token": "123", "authorization": {}}`,
	}
	withTokenNewer := types.ExternalService{
		ID:          4,
		Kind:        extsvc.KindGitHub,
		DisplayName: "GitHub newer token",
		Config:      `{"url": "https://github.com", "token": "123456", "authorization": {}}`,
	}

	repo := &types.Repo{
		Name:    api.RepoName("test-repo"),
		URI:     "test-repo",
		Private: true,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "external-id-123",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com/",
		},
		Sources: map[string]*types.SourceInfo{
			noToken.URN(): {
				ID:       noToken.URN(),
				CloneURL: "https://github.com/sourcegraph/sourcegraph",
			},
			userOwnedWithToken.URN(): {
				ID:       userOwnedWithToken.URN(),
				CloneURL: "https://123@github.com/sourcegraph/sourcegraph",
			},
			withToken.URN(): {
				ID:       withToken.URN(),
				CloneURL: "https://123@github.com/sourcegraph/sourcegraph",
			},
			withTokenNewer.URN(): {
				ID:       withTokenNewer.URN(),
				CloneURL: "https://123456@github.com/sourcegraph/sourcegraph",
			},
		},
	}

	database.Mocks.ExternalServices.List = func(opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
		sources := make([]*types.ExternalService, 0)
		if _, ok := repo.Sources[noToken.URN()]; ok {
			sources = append(sources, &noToken)
		}
		if _, ok := repo.Sources[userOwnedWithToken.URN()]; ok {
			sources = append(sources, &userOwnedWithToken)
		}
		if _, ok := repo.Sources[withToken.URN()]; ok {
			sources = append(sources, &withToken)
		}
		if _, ok := repo.Sources[withTokenNewer.URN()]; ok {
			sources = append(sources, &withTokenNewer)
		}
		return sources, nil
	}
	t.Cleanup(func() {
		database.Mocks.ExternalServices.List = nil
	})

	// Expect the newest public external service with a token to be returned.
	svc, err := loadExternalService(ctx, &MockSyncStore{}, repo)
	if err != nil {
		t.Fatalf("invalid error, expected nil, got %v", err)
	}
	if have, want := svc.ID, withTokenNewer.ID; have != want {
		t.Fatalf("invalid external service returned, want=%d have=%d", want, have)
	}

	// Now delete the global external services and expect the user owned external service to be returned.
	delete(repo.Sources, withTokenNewer.URN())
	delete(repo.Sources, withToken.URN())
	svc, err = loadExternalService(ctx, &MockSyncStore{}, repo)
	if err != nil {
		t.Fatalf("invalid error, expected nil, got %v", err)
	}
	if have, want := svc.ID, userOwnedWithToken.ID; have != want {
		t.Fatalf("invalid external service returned, want=%d have=%d", want, have)
	}
}

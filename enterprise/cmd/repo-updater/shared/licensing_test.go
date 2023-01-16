package shared

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestEnterpriseCreateRepoHook(t *testing.T) {
	ctx := context.Background()

	// Set up mock repo count
	mockRepoStore := database.NewMockRepoStore()
	mockStore := repos.NewMockStore()
	mockStore.RepoStoreFunc.SetDefaultReturn(mockRepoStore)

	tests := map[string]*struct {
		maxPrivateRepos int
		unrestricted    bool
		numPrivateRepos int
		newRepo         *types.Repo
		wantErr         bool
	}{
		"private repo, unrestricted": {
			unrestricted:    true,
			numPrivateRepos: 99999999,
			newRepo:         &types.Repo{Private: true},
			wantErr:         false,
		},
		"private repo, max private repos reached": {
			maxPrivateRepos: 1,
			numPrivateRepos: 1,
			newRepo:         &types.Repo{Private: true},
			wantErr:         true,
		},
		"public repo, max private repos reached": {
			maxPrivateRepos: 1,
			numPrivateRepos: 1,
			newRepo:         &types.Repo{Private: false},
			wantErr:         false,
		},
		"private repo, max private repos not reached": {
			maxPrivateRepos: 2,
			numPrivateRepos: 1,
			newRepo:         &types.Repo{Private: true},
			wantErr:         false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockRepoStore.CountFunc.SetDefaultReturn(test.numPrivateRepos, nil)

			defaultMock := licensing.MockCheckFeature
			licensing.MockCheckFeature = func(feature licensing.Feature) error {
				if prFeature, ok := feature.(*licensing.FeaturePrivateRepositories); ok {
					prFeature.MaxNumPrivateRepos = test.maxPrivateRepos
					prFeature.Unrestricted = test.unrestricted
				}

				return nil
			}
			defer func() {
				licensing.MockCheckFeature = defaultMock
			}()

			err := enterpriseCreateRepoHook(ctx, mockStore, test.newRepo)
			if gotErr := err != nil; gotErr != test.wantErr {
				t.Fatalf("got err: %t, want err: %t, err: %q", gotErr, test.wantErr, err)
			}
		})
	}
}

func TestEnterpriseUpdateRepoHook(t *testing.T) {
	ctx := context.Background()

	// Set up mock repo count
	mockRepoStore := database.NewMockRepoStore()
	mockStore := repos.NewMockStore()
	mockStore.RepoStoreFunc.SetDefaultReturn(mockRepoStore)

	tests := map[string]*struct {
		maxPrivateRepos int
		unrestricted    bool
		numPrivateRepos int
		existingRepo    *types.Repo
		newRepo         *types.Repo
		wantErr         bool
	}{
		"from public to private, unrestricted": {
			unrestricted:    true,
			numPrivateRepos: 99999999,
			existingRepo:    &types.Repo{Private: false},
			newRepo:         &types.Repo{Private: true},
			wantErr:         false,
		},
		"from public to private, max private repos reached": {
			maxPrivateRepos: 1,
			numPrivateRepos: 1,
			existingRepo:    &types.Repo{Private: false},
			newRepo:         &types.Repo{Private: true},
			wantErr:         true,
		},
		"from private to private, max private repos reached": {
			maxPrivateRepos: 1,
			numPrivateRepos: 1,
			existingRepo:    &types.Repo{Private: true},
			newRepo:         &types.Repo{Private: true},
			wantErr:         false,
		},
		"from private to public, max private repos reached": {
			maxPrivateRepos: 1,
			numPrivateRepos: 1,
			existingRepo:    &types.Repo{Private: true},
			newRepo:         &types.Repo{Private: false},
			wantErr:         false,
		},
		"from public to private, max private repos not reached": {
			maxPrivateRepos: 2,
			numPrivateRepos: 1,
			existingRepo:    &types.Repo{Private: false},
			newRepo:         &types.Repo{Private: true},
			wantErr:         false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockRepoStore.CountFunc.SetDefaultReturn(test.numPrivateRepos, nil)

			defaultMock := licensing.MockCheckFeature
			licensing.MockCheckFeature = func(feature licensing.Feature) error {
				if prFeature, ok := feature.(*licensing.FeaturePrivateRepositories); ok {
					prFeature.MaxNumPrivateRepos = test.maxPrivateRepos
					prFeature.Unrestricted = test.unrestricted
				}

				return nil
			}
			defer func() {
				licensing.MockCheckFeature = defaultMock
			}()

			err := enterpriseUpdateRepoHook(ctx, mockStore, test.existingRepo, test.newRepo)
			if gotErr := err != nil; gotErr != test.wantErr {
				t.Fatalf("got err: %t, want err: %t, err: %q", gotErr, test.wantErr, err)
			}
		})
	}
}

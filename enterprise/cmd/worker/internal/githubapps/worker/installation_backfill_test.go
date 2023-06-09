package worker

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-github/v41/github"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/github_apps/store"
	ghtypes "github.com/sourcegraph/sourcegraph/enterprise/internal/github_apps/types"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type mockGitHubClient struct {
	mock.Mock
}

func (m *mockGitHubClient) GetAppInstallations(ctx context.Context) ([]*github.Installation, error) {
	args := m.Called(ctx)
	if args.Get(0) != nil {
		return args.Get(0).([]*github.Installation), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestGitHubInstallationWorker(t *testing.T) {
	apps := []*ghtypes.GitHubApp{
		{
			ID:         1,
			Name:       "Github App One",
			AppID:      10001,
			BaseURL:    "https://example.com",
			PrivateKey: "random-private-key",
		},
		{
			ID:         2,
			Name:       "Github App Two",
			AppID:      10002,
			BaseURL:    "https://example.com",
			PrivateKey: "random-private-key",
		},
		{
			ID:         3,
			Name:       "Github App Three",
			AppID:      10003,
			BaseURL:    "https://example.com",
			PrivateKey: "random-private-key",
		},
	}
	ghStore := store.NewMockGitHubAppsStore()
	ghStore.ListFunc.SetDefaultHook(func(ctx context.Context, ghad *types.GitHubAppDomain) ([]*ghtypes.GitHubApp, error) {
		return apps, nil
	})

	// The assumption here is that:
	// GitHub app with ID 1 has no installation of it saved in the database
	// GitHub app with ID 2 has one installation of it saved in the database
	// GitHub app with ID 3 has two installations of it saved in the database
	ghStore.GetInstallationsFunc.SetDefaultHook(func(ctx context.Context, i int) ([]*ghtypes.GitHubAppInstallation, error) {
		installs := []*ghtypes.GitHubAppInstallation{}

		if i == 2 {
			installs = append(installs, &ghtypes.GitHubAppInstallation{
				ID:             1,
				AppID:          i,
				InstallationID: 120400,
			})
		}

		if i == 3 {
			installs = append(installs, &ghtypes.GitHubAppInstallation{
				ID:             2,
				AppID:          i,
				InstallationID: 120402,
			}, &ghtypes.GitHubAppInstallation{
				ID:             3,
				AppID:          i,
				InstallationID: 120403,
			})
		}

		return installs, nil
	})

	ghStore.InstallFunc.SetDefaultHook(func(ctx context.Context, in ghtypes.GitHubAppInstallation) (*ghtypes.GitHubAppInstallation, error) {
		fmt.Println("install: appID: ", in.AppID, " installationID: ", in.InstallationID)
		return nil, nil
	})
	ghStore.BulkRemoveInstallationsFunc.SetDefaultHook(func(ctx context.Context, i1 int, i2 []int) error {
		fmt.Println("bulk remove: app.ID: ", i1, " installations.ID: ", i2)
		return nil
	})

	db := database.NewMockEnterpriseDB()
	db.GitHubAppsFunc.SetDefaultReturn(ghStore)

	logger := logtest.Scoped(t)
	worker := NewGitHubInstallationWorker(db, logger)
	ctx := context.Background()

	MockGitHubClient = func(app *ghtypes.GitHubApp, logger log.Logger) (GitHubAppClient, error) {
		client := new(mockGitHubClient)
		if app.ID == 1 {
			client.On("GetAppInstallations", mock.Anything).Return([]*github.Installation{
				{
					ID: github.Int64(120398),
				},
				{
					ID: github.Int64(120399),
				},
			}, nil)
		}

		if app.ID == 2 {
			client.On("GetAppInstallations", mock.Anything).Return([]*github.Installation{
				{
					ID: github.Int64(120400),
				},
				{
					ID: github.Int64(120401),
				},
			}, nil)
		}

		if app.ID == 3 {
			client.On("GetAppInstallations", mock.Anything).Return([]*github.Installation{
				{
					ID: github.Int64(120402),
				},
			}, nil)
		}

		return client, nil
	}

	err := worker.Handle(ctx)
	require.NoError(t, err)

	// We bulk install two installations for GitHub app with ID 1 and one installation for ID 2
	if len(ghStore.InstallFunc.History()) != 3 {
		t.Errorf("expected 3 calls to Install, got %d", len(ghStore.InstallFunc.History()))
	}

	// We bulk remove on3 installation for GitHub app with ID 3
	if len(ghStore.BulkRemoveInstallationsFunc.History()) != 1 {
		t.Errorf("expected 1 call to BulkRemove, got %d", len(ghStore.BulkRemoveInstallationsFunc.History()))
	}
}

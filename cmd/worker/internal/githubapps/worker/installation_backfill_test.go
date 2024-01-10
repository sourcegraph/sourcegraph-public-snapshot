package worker

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-github/v55/github"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/github_apps/store"
	ghtypes "github.com/sourcegraph/sourcegraph/internal/github_apps/types"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
	ghStore := store.NewMockGitHubAppsStore()

	apps := []*ghtypes.GitHubApp{{ID: 1, AppID: 1}, {ID: 2, AppID: 2}, {ID: 3, AppID: 3}}
	ghStore.ListFunc.SetDefaultHook(func(ctx context.Context, ghad *types.GitHubAppDomain) ([]*ghtypes.GitHubApp, error) {
		return apps, nil
	})
	ghStore.SyncInstallationsFunc.SetDefaultHook(func(ctx context.Context, app ghtypes.GitHubApp, logger log.Logger, client ghtypes.GitHubAppClient) (errs errors.MultiError) {
		fmt.Println("sync installations: ", app.ID)
		return nil
	})

	db := dbmocks.NewMockDB()
	db.GitHubAppsFunc.SetDefaultReturn(ghStore)

	logger := logtest.Scoped(t)
	worker := NewGitHubInstallationWorker(db, logger)
	ctx := context.Background()

	MockGitHubClient = func(app *ghtypes.GitHubApp, logger log.Logger) (ghtypes.GitHubAppClient, error) {
		return new(mockGitHubClient), nil
	}

	err := worker.Handle(ctx)
	require.NoError(t, err)

	// We upsert all installations we received
	if len(ghStore.SyncInstallationsFunc.History()) != 3 {
		t.Errorf("expected 3 calls to SyncInstallations, got %d", len(ghStore.SyncInstallationsFunc.History()))
	}
}

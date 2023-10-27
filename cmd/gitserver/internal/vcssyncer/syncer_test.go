package vcssyncer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	api "github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestGetVCSSyncer(t *testing.T) {
	tempDir := t.TempDir()
	tempReposDir, err := os.MkdirTemp(tempDir, "TestGetVCSSyncer")
	if err != nil {
		t.Fatal(err)
	}
	tempCoursierCacheDir := filepath.Join(tempReposDir, "coursier")

	repo := api.RepoName("foo/bar")
	extsvcStore := dbmocks.NewMockExternalServiceStore()
	repoStore := dbmocks.NewMockRepoStore()

	repoStore.GetByNameFunc.SetDefaultHook(func(ctx context.Context, name api.RepoName) (*types.Repo, error) {
		return &types.Repo{
			ExternalRepo: api.ExternalRepoSpec{
				ServiceType: extsvc.TypePerforce,
			},
			Sources: map[string]*types.SourceInfo{
				"a": {
					ID:       "abc",
					CloneURL: "example.com",
				},
			},
		}, nil
	})

	extsvcStore.GetByIDFunc.SetDefaultHook(func(ctx context.Context, i int64) (*types.ExternalService, error) {
		return &types.ExternalService{
			ID:          1,
			Kind:        extsvc.KindPerforce,
			DisplayName: "test",
			Config:      extsvc.NewEmptyConfig(),
		}, nil
	})

	s, err := NewVCSSyncer(context.Background(), &NewVCSSyncerOpts{
		ExternalServiceStore: extsvcStore,
		RepoStore:            repoStore,
		DepsSvc:              new(dependencies.Service),
		Repo:                 repo,
		ReposDir:             tempReposDir,
		CoursierCacheDir:     tempCoursierCacheDir,
		Logger:               logtest.Scoped(t),
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, "perforce", s.Type())
}

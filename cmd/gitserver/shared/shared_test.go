package shared

import (
	"context"
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		logtest.InitWithLevel(m, log.LevelNone)
	}
	os.Exit(m.Run())
}

func TestParsePercent(t *testing.T) {
	tests := []struct {
		i       int
		want    int
		wantErr bool
	}{
		{i: -1, wantErr: true},
		{i: -4, wantErr: true},
		{i: 300, wantErr: true},
		{i: 0, want: 0},
		{i: 50, want: 50},
		{i: 100, want: 100},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got, err := getPercent(tt.i)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePercent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parsePercent() = %v, want %v", got, tt.want)
			}
		})
	}
}

type mockDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (c *mockDoer) Do(r *http.Request) (*http.Response, error) {
	return c.do(r)
}

func TestGetVCSSyncer(t *testing.T) {
	tempReposDir, err := os.MkdirTemp("", "TestGetVCSSyncer")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempReposDir); err != nil {
			t.Fatal(err)
		}
	})
	tempCoursierCacheDir := filepath.Join(tempReposDir, "coursier")

	repo := api.RepoName("foo/bar")
	extsvcStore := database.NewMockExternalServiceStore()
	repoStore := database.NewMockRepoStore()
	depsSvc := new(dependencies.Service)

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

	s, err := getVCSSyncer(context.Background(), extsvcStore, repoStore, depsSvc, repo, tempReposDir, tempCoursierCacheDir)
	if err != nil {
		t.Fatal(err)
	}

	_, ok := s.(*server.PerforceDepotSyncer)
	if !ok {
		t.Fatalf("Want *server.PerforceDepotSyncer, got %T", s)
	}
}

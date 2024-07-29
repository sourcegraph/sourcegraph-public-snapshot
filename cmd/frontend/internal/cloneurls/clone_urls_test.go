package cloneurls

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestReposourceCloneURLToRepoName(t *testing.T) {
	ctx := context.Background()

	externalServices := dbmocks.NewMockExternalServiceStore()
	externalServices.ListFunc.SetDefaultReturn(
		[]*types.ExternalService{
			{
				ID:          1,
				Kind:        extsvc.KindGitHub,
				DisplayName: "GITHUB #1",
				Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.example.com", "repositoryQuery": ["none"], "token": "abc"}`),
			},
			{
				ID:          2,
				Kind:        extsvc.KindGerrit,
				DisplayName: "GERRIT #1",
				Config:      extsvc.NewUnencryptedConfig(`{"url": "https://gerrit.example.com"}`),
			},
		},
		nil,
	)

	db := dbmocks.NewMockDB()
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)
	db.ReposFunc.SetDefaultReturn(dbmocks.NewMockRepoStore())

	tests := []struct {
		name         string
		cloneURL     string
		wantRepoName api.RepoName
	}{
		{
			name:     "no match",
			cloneURL: "https://gitlab.com/user/repo",
		},
		{
			name:         "match existing external service",
			cloneURL:     "https://github.example.com/user/repo.git",
			wantRepoName: api.RepoName("github.example.com/user/repo"),
		},
		{
			name:         "fallback for github.com",
			cloneURL:     "https://github.com/user/repo",
			wantRepoName: api.RepoName("github.com/user/repo"),
		},
		{
			name:         "relatively-pathed submodule",
			cloneURL:     "../../a/b/c.git",
			wantRepoName: api.RepoName("github.example.com/a/b/c"),
		},
		{
			name:         "gerrit",
			cloneURL:     "https://gerrit.example.com/a/repo.git",
			wantRepoName: api.RepoName("gerrit.example.com/repo"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repoName, err := RepoSourceCloneURLToRepoName(ctx, db, test.cloneURL)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.wantRepoName, repoName); diff != "" {
				t.Fatalf("RepoName mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

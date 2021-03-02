package database

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestIterateRepoGitserverStatus(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	db := dbtesting.GetDB(t)
	ctx := context.Background()

	repo1 := &types.Repo{
		Name:         "github.com/sourcegraph/repo1",
		URI:          "github.com/sourcegraph/repo1",
		Description:  "",
		ExternalRepo: api.ExternalRepoSpec{},
		Sources:      nil,
	}
	repo2 := &types.Repo{
		Name:         "github.com/sourcegraph/repo2",
		URI:          "github.com/sourcegraph/repo2",
		Description:  "",
		ExternalRepo: api.ExternalRepoSpec{},
		Sources:      nil,
	}

	// Create two test repos
	err := Repos(db).Create(ctx, repo1, repo2)
	if err != nil {
		t.Fatal(err)
	}

	gitserverRepo := &types.GitserverRepo{
		RepoID:              repo1.ID,
		ShardID:             "gitserver1",
		CloneStatus:         "not_cloned",
		LastExternalService: 0,
	}

	// Create one GitServerRepo
	if err := GitserverRepos(db).Create(ctx, gitserverRepo); err != nil {
		t.Fatal(err)
	}

	var repoCount int
	var statusCount int

	// Iterate
	err = GitserverRepos(db).IterateRepoGitserverStatus(ctx, func(repo types.RepoGitserverStatus) error {
		repoCount++
		if repo.GitserverRepo != nil {
			statusCount++
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	wantRepoCount := 2
	if repoCount != wantRepoCount {
		t.Fatalf("Expected %d repos, got %d", wantRepoCount, repoCount)
	}

	wantStatusCount := 1
	if statusCount != wantStatusCount {
		t.Fatalf("Expected %d statuses, got %d", wantStatusCount, statusCount)
	}
}

func TestGitserverRepoCreateNullShard(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	db := dbtesting.GetDB(t)
	ctx := context.Background()

	repo1 := &types.Repo{
		Name:         "github.com/sourcegraph/repo1",
		URI:          "github.com/sourcegraph/repo1",
		Description:  "",
		ExternalRepo: api.ExternalRepoSpec{},
		Sources:      nil,
	}

	// Create two test repos
	err := Repos(db).Create(ctx, repo1)
	if err != nil {
		t.Fatal(err)
	}

	gitserverRepo := &types.GitserverRepo{
		RepoID:              repo1.ID,
		ShardID:             "",
		CloneStatus:         "not_cloned",
		LastExternalService: 0,
	}

	// Create one GitServerRepo
	if err := GitserverRepos(db).Create(ctx, gitserverRepo); err == nil {
		t.Fatal("Expected an error")
	}
}

package db

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestUserPublicRepos_Set(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	user, err := Users.Create(ctx, NewUser{
		Username: "u",
		Password: "p",
	})
	if err != nil {
		t.Errorf("Expected no error, got %s ", err)
	}

	err = Repos.Create(ctx, &types.Repo{
		Name: "test",
	})
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	repo, err := Repos.GetByName(ctx, "test")
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	err = UserPublicRepos.SetUserRepo(ctx, user.ID, repo.ID)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	repoIDs, err := UserPublicRepos.ListByUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	if wanted, got := 1, len(repoIDs); wanted != got {
		t.Errorf("wanted %v repos, got %v", wanted, got)
	}
	if wanted, got := int32(repo.ID), repoIDs[0]; wanted != got {
		t.Errorf("wanted repo ID %v, got %v", wanted, got)
	}
}

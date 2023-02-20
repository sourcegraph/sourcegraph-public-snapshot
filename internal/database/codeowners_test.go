package database

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/v1"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestCodeowners_CreateUpdateDelete(t *testing.T) {
	ctx := context.Background()

	logger := logtest.NoOp(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	store := db.Codeowners()

	t.Run("create new codeowners file", func(t *testing.T) {
		codeowners := newCodeownersFile("*", "everyone", api.RepoID(100))
		if err := store.CreateCodeownersFile(ctx, codeowners); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("create codeowners duplicate errors", func(t *testing.T) {
		codeowners := newCodeownersFile("*", "everyone", api.RepoID(200))
		if err := store.CreateCodeownersFile(ctx, codeowners); err != nil {
			t.Fatal(err)
		}
		secondErr := store.CreateCodeownersFile(ctx, codeowners)
		if secondErr == nil {
			t.Fatal("expect duplicate codeowners to error")
		}
		require.ErrorAs(t, ErrCodeownersFileAlreadyExists, &secondErr)
	})

	t.Run("update codeowners file", func(t *testing.T) {
		codeowners := newCodeownersFile("*", "everyone", api.RepoID(102))
		if err := store.CreateCodeownersFile(ctx, codeowners); err != nil {
			t.Fatal(err)
		}
		codeowners = newCodeownersFile("*", "notEveryone", api.RepoID(102))
		if err := store.UpdateCodeownersFile(ctx, codeowners); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("update non existent codeowners file", func(t *testing.T) {
		codeowners := newCodeownersFile("*", "notEveryone", api.RepoID(33))
		err := store.UpdateCodeownersFile(ctx, codeowners)
		if err == nil {
			t.Fatal("expected not found error")
		}
		require.ErrorAs(t, CodeownersFileNotFoundError{}, &err)
	})

	t.Run("delete", func(t *testing.T) {
		repoID := api.RepoID(10)
		codeowners := newCodeownersFile("*", "everyone", repoID)
		if err := store.CreateCodeownersFile(ctx, codeowners); err != nil {
			t.Fatal(err)
		}
		if err := store.DeleteCodeownersForRepo(ctx, repoID); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("delete non existent codeowners file", func(t *testing.T) {
		err := store.DeleteCodeownersForRepo(ctx, api.RepoID(9000))
		if err == nil {
			t.Fatal("did not return useful not found information")
		}
		require.ErrorAs(t, CodeownersFileNotFoundError{}, &err)
	})
}

func TestCodeowners_Get(t *testing.T) {
	ctx := context.Background()

	logger := logtest.NoOp(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	store := db.Codeowners()

	t.Run("not found", func(t *testing.T) {
		_, err := store.GetCodeownersForRepo(ctx, api.RepoID(100))
		if err == nil {
			t.Fatal("expected an error")
		}
		require.ErrorAs(t, CodeownersFileNotFoundError{}, &err)
	})

	t.Run("get by repo ID", func(t *testing.T) {
		codeownersFile := newCodeownersFile("*", "person", api.RepoID(11))
		if err := store.CreateCodeownersFile(ctx, codeownersFile); err != nil {
			t.Fatal("creating codeowners failed", err)
		}
		got, err := store.GetCodeownersForRepo(ctx, api.RepoID(11))
		if err != nil {
			t.Fatal(err)
		}
		require.Equal(t, codeownersFile, got)
	})

	t.Run("get by repo ID after update", func(t *testing.T) {
		codeownersFile := newCodeownersFile("*", "everyone", api.RepoID(102))
		if err := store.CreateCodeownersFile(ctx, codeownersFile); err != nil {
			t.Fatal(err)
		}
		got, err := store.GetCodeownersForRepo(ctx, api.RepoID(102))
		if err != nil {
			t.Fatal(err)
		}
		require.Equal(t, codeownersFile, got)
		codeownersFile.UpdatedAt = time.Now().UTC()
		if err := store.UpdateCodeownersFile(ctx, codeownersFile); err != nil {
			t.Fatal(err)
		}
		got, err = store.GetCodeownersForRepo(ctx, api.RepoID(102))
		if err != nil {
			t.Fatal(err)
		}
		require.Equal(t, codeownersFile, got)
	})
}

// newCodeownersFile returns a simple test Codeowners file with one pattern and one owner.
func newCodeownersFile(pattern, handle string, repoID api.RepoID) *types.CodeownersFile {
	return &types.CodeownersFile{
		Contents: fmt.Sprintf("%s @%s", pattern, handle),
		Proto: &codeownerspb.File{
			Rule: []*codeownerspb.Rule{
				{
					Pattern: pattern,
					Owner:   []*codeownerspb.Owner{{Handle: handle}},
				},
			},
		},
		RepoID: repoID,
	}
}

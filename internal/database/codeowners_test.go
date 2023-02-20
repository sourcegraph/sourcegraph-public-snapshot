package database

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
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

	t.Run("create codeowners duplicate error", func(t *testing.T) {
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

func TestCodeowners_GetList(t *testing.T) {
	ctx := context.Background()

	logger := logtest.NoOp(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	store := db.Codeowners()

	createFile := func(file *types.CodeownersFile) *types.CodeownersFile {
		if err := store.CreateCodeownersFile(ctx, file); err != nil {
			t.Fatal(err)
		}
		return file
	}

	repo11Codeowners := createFile(newCodeownersFile("*", "person", api.RepoID(11)))
	repo102Codeowners := createFile(newCodeownersFile("*", "everyone", api.RepoID(102)))

	t.Run("get", func(t *testing.T) {
		t.Run("not found", func(t *testing.T) {
			_, err := store.GetCodeownersForRepo(ctx, api.RepoID(100))
			if err == nil {
				t.Fatal("expected an error")
			}
			require.ErrorAs(t, CodeownersFileNotFoundError{}, &err)
		})
		t.Run("get by repo ID", func(t *testing.T) {
			got, err := store.GetCodeownersForRepo(ctx, api.RepoID(11))
			if err != nil {
				t.Fatal(err)
			}
			require.Equal(t, repo11Codeowners, got)
		})
		t.Run("get by repo ID after update", func(t *testing.T) {
			got, err := store.GetCodeownersForRepo(ctx, api.RepoID(102))
			if err != nil {
				t.Fatal(err)
			}
			require.Equal(t, repo102Codeowners, got)
			repo102Codeowners.UpdatedAt = time.Now().UTC()
			if err := store.UpdateCodeownersFile(ctx, repo102Codeowners); err != nil {
				t.Fatal(err)
			}
			got, err = store.GetCodeownersForRepo(ctx, api.RepoID(102))
			if err != nil {
				t.Fatal(err)
			}
			require.Equal(t, repo102Codeowners, got)
		})
	})

	t.Run("list", func(t *testing.T) {
		all := []*types.CodeownersFile{repo11Codeowners, repo102Codeowners}

		// List all
		have, cursor, err := store.ListCodeowners(ctx, ListCodeownersOpts{})
		if err != nil {
			t.Fatal(err)
		}
		require.Equal(t, all, have)
		if cursor != 0 {
			t.Fatal("incorrect cursor returned")
		}

		// List with cursor pagination
		var lastCursor int32
		for i := 0; i < len(all); i++ {
			t.Run(fmt.Sprintf("list codeowners n#%d", i), func(t *testing.T) {
				opts := ListCodeownersOpts{LimitOffset: &LimitOffset{Limit: 1}, Cursor: lastCursor}
				cf, c, err := store.ListCodeowners(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}
				lastCursor = c
				fmt.Println(cf[0])
				assert.Equal(t, all[i], cf[0])
			})
		}
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

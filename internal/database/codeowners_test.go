package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/v1"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestCodeowners_CreateUpdate(t *testing.T) {
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

	t.Run("update codeowners file", func(t *testing.T) {

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

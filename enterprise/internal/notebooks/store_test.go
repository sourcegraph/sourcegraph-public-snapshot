package notebooks

import (
	"context"
	"reflect"
	"testing"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func createNotebooks(ctx context.Context, store NotebooksStore, notebooks []*Notebook) ([]*Notebook, error) {
	createdNotebooks := make([]*Notebook, len(notebooks))
	for idx, notebook := range notebooks {
		createdNotebook, err := store.CreateNotebook(ctx, notebook)
		if err != nil {
			return nil, err
		}
		createdNotebooks[idx] = createdNotebook
	}
	return createdNotebooks, nil
}

func TestCreateAndGetNotebook(t *testing.T) {
	t.Parallel()
	db := dbtest.NewDB(t)
	ctx := actor.WithInternalActor(context.Background())
	u := database.Users(db)
	n := Notebooks(db)

	user, err := u.Create(ctx, database.NewUser{Username: "u", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	blocks := NotebookBlocks{
		{ID: "1", Type: NotebookQueryBlockType, QueryInput: &NotebookQueryBlockInput{"repo:a b"}},
		{ID: "2", Type: NotebookMarkdownBlockType, MarkdownInput: &NotebookMarkdownBlockInput{"# Title"}},
		{ID: "3", Type: NotebookFileBlockType, FileInput: &NotebookFileBlockInput{
			RepositoryName: "github.com/sourcegraph/sourcegraph", FilePath: "client/web/file.tsx"},
		},
	}
	notebook := &Notebook{Title: "Notebook Title", Blocks: blocks, Public: true, CreatorUserID: user.ID}
	createdNotebook, err := n.CreateNotebook(ctx, notebook)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(blocks, createdNotebook.Blocks) {
		t.Fatalf("wanted %v blocks, got %v", blocks, createdNotebook.Blocks)
	}
}

func TestCreatingNotebookWithInvalidBlock(t *testing.T) {
	t.Parallel()
	db := dbtest.NewDB(t)
	ctx := actor.WithInternalActor(context.Background())
	u := database.Users(db)
	n := Notebooks(db)

	user, err := u.Create(ctx, database.NewUser{Username: "u", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	blocks := NotebookBlocks{{ID: "1", Type: NotebookQueryBlockType}}
	notebook := &Notebook{Title: "Notebook Title", Blocks: blocks, Public: true, CreatorUserID: user.ID}
	_, err = n.CreateNotebook(ctx, notebook)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	wantErr := "invalid query block with id: 1"
	if err.Error() != wantErr {
		t.Fatalf("wanted '%s' error, got '%s'", wantErr, err.Error())
	}
}

func TestNotebookPermissions(t *testing.T) {
	t.Parallel()
	db := dbtest.NewDB(t)
	internalCtx := actor.WithInternalActor(context.Background())
	u := database.Users(db)
	n := Notebooks(db)

	user1, err := u.Create(internalCtx, database.NewUser{Username: "u1", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	user2, err := u.Create(internalCtx, database.NewUser{Username: "u2", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	createdNotebooks, err := createNotebooks(internalCtx, n, []*Notebook{
		{Title: "Notebook User1 Public", Blocks: NotebookBlocks{}, Public: true, CreatorUserID: user1.ID},
		{Title: "Notebook User1 Private", Blocks: NotebookBlocks{}, Public: false, CreatorUserID: user1.ID},
	})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		notebookID int64
		userID     int32
		wantErr    *error
	}{
		{notebookID: createdNotebooks[0].ID, userID: user1.ID, wantErr: nil},
		{notebookID: createdNotebooks[1].ID, userID: user1.ID, wantErr: nil},
		// User2 *can* access a public notebook from a different user (User1)
		{notebookID: createdNotebooks[0].ID, userID: user2.ID, wantErr: nil},
		// User2 *cannot* access a private notebook from a different user (User1)
		{notebookID: createdNotebooks[1].ID, userID: user2.ID, wantErr: &ErrNotebookNotFound},
	}

	for _, tt := range tests {
		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: tt.userID})
		_, err := n.GetNotebook(ctx, tt.notebookID)
		if tt.wantErr != nil && !errors.Is(err, *tt.wantErr) {
			t.Errorf("expected error not found in chain: got %+v, want %+v", err, *tt.wantErr)
		} else if tt.wantErr == nil && err != nil {
			t.Errorf("expected no error, got %+v", err)
		}
	}
}

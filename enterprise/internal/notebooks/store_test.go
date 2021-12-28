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

func TestUpdateNotebook(t *testing.T) {
	t.Parallel()
	db := dbtest.NewDB(t)
	ctx := actor.WithInternalActor(context.Background())
	u := database.Users(db)
	n := Notebooks(db)

	user, err := u.Create(ctx, database.NewUser{Username: "u", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	blocks := NotebookBlocks{{ID: "1", Type: NotebookQueryBlockType, QueryInput: &NotebookQueryBlockInput{"repo:a b"}}}
	notebook := &Notebook{Title: "Notebook Title", Blocks: blocks, Public: true, CreatorUserID: user.ID}
	createdNotebook, err := n.CreateNotebook(ctx, notebook)
	if err != nil {
		t.Fatal(err)
	}

	wantUpdatedNotebook := createdNotebook
	wantUpdatedNotebook.Title = "Notebook Title 1"
	wantUpdatedNotebook.Public = false
	wantUpdatedNotebook.Blocks = NotebookBlocks{{ID: "2", Type: NotebookMarkdownBlockType, MarkdownInput: &NotebookMarkdownBlockInput{"# Title"}}}

	gotUpdatedNotebook, err := n.UpdateNotebook(ctx, wantUpdatedNotebook)
	if err != nil {
		t.Fatal(err)
	}

	// Ignore updatedAt change
	wantUpdatedNotebook.UpdatedAt = gotUpdatedNotebook.UpdatedAt

	if !reflect.DeepEqual(wantUpdatedNotebook, gotUpdatedNotebook) {
		t.Fatalf("wanted %+v updated notebook, got %+v", wantUpdatedNotebook, gotUpdatedNotebook)
	}
}

func TestDeleteNotebook(t *testing.T) {
	t.Parallel()
	db := dbtest.NewDB(t)
	ctx := actor.WithInternalActor(context.Background())
	u := database.Users(db)
	n := Notebooks(db)

	user, err := u.Create(ctx, database.NewUser{Username: "u", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	blocks := NotebookBlocks{{ID: "1", Type: NotebookQueryBlockType, QueryInput: &NotebookQueryBlockInput{"repo:a b"}}}
	notebook := &Notebook{Title: "Notebook Title", Blocks: blocks, Public: true, CreatorUserID: user.ID}
	createdNotebook, err := n.CreateNotebook(ctx, notebook)
	if err != nil {
		t.Fatal(err)
	}

	err = n.DeleteNotebook(ctx, createdNotebook.ID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = n.GetNotebook(ctx, createdNotebook.ID)
	if !errors.Is(err, ErrNotebookNotFound) {
		t.Fatalf("want ErrNotebookNotFound error, got %+v", err)
	}
}

func getNotebookIDs(notebooks []*Notebook) []int64 {
	ids := make([]int64, 0, len(notebooks))
	for _, n := range notebooks {
		ids = append(ids, n.ID)
	}
	return ids
}

func TestConvertingToPostgresTextSearchQuery(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		wantTSQuery string
	}{
		{
			name:        "single token",
			query:       "asimplequery",
			wantTSQuery: "asimplequery:*",
		},
		{
			name:        "multiple tokens",
			query:       "a simple query",
			wantTSQuery: "a:* & simple:* & query:*",
		},
		{
			name:        "special chars",
			query:       "a & special | q:u !e (r y)",
			wantTSQuery: "a:* & special:* & q:* & u:* & e:* & r:* & y:*",
		},
	}

	for _, tt := range tests {
		gotTSQuery := toPostgresTextSearchQuery(tt.query)
		if tt.wantTSQuery != gotTSQuery {
			t.Fatalf("wanted '%s' text search query, got '%s'", tt.wantTSQuery, gotTSQuery)
		}
	}
}

func TestListingAndCountingNotebooks(t *testing.T) {
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

	blocks := NotebookBlocks{
		{ID: "1", Type: NotebookQueryBlockType, QueryInput: &NotebookQueryBlockInput{"repo:a b"}},
		{ID: "2", Type: NotebookMarkdownBlockType, MarkdownInput: &NotebookMarkdownBlockInput{"# Title"}},
		{ID: "3", Type: NotebookQueryBlockType, QueryInput: &NotebookQueryBlockInput{"repo:sourcegraph file:client/web/file.tsx"}},
		{ID: "4", Type: NotebookFileBlockType, FileInput: &NotebookFileBlockInput{
			RepositoryName: "github.com/sourcegraph/sourcegraph", FilePath: "client/web/file.tsx"},
		},
		{ID: "5", Type: NotebookMarkdownBlockType, MarkdownInput: &NotebookMarkdownBlockInput{"Lorem ipsum dolor sit amet, consectetur adipiscing elit."}},
		{ID: "6", Type: NotebookMarkdownBlockType, MarkdownInput: &NotebookMarkdownBlockInput{"Donec in auctor odio."}},
	}

	createdNotebooks, err := createNotebooks(internalCtx, n, []*Notebook{
		{Title: "Notebook User1 Public", Blocks: NotebookBlocks{blocks[0], blocks[4]}, Public: true, CreatorUserID: user1.ID},
		{Title: "Notebook User1 Private", Blocks: NotebookBlocks{blocks[1]}, Public: false, CreatorUserID: user1.ID},
		{Title: "Notebook User2 Public", Blocks: NotebookBlocks{blocks[2], blocks[5]}, Public: true, CreatorUserID: user2.ID},
		{Title: "Notebook User2 Private", Blocks: NotebookBlocks{blocks[3]}, Public: false, CreatorUserID: user2.ID},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = n.UpdateNotebook(internalCtx, createdNotebooks[0])
	if err != nil {
		t.Fatal(err)
	}
	_, err = n.UpdateNotebook(internalCtx, createdNotebooks[2])
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name            string
		userID          int32
		pageOpts        ListNotebooksPageOptions
		opts            ListNotebooksOptions
		wantNotebookIDs []int64
		wantCount       int64
	}{
		{
			name:            "get all user1 accessible notebooks",
			userID:          user1.ID,
			pageOpts:        ListNotebooksPageOptions{First: 4},
			opts:            ListNotebooksOptions{},
			wantNotebookIDs: getNotebookIDs(createdNotebooks[:3]),
			wantCount:       3,
		},
		{
			name:            "get notebooks page",
			userID:          user1.ID,
			pageOpts:        ListNotebooksPageOptions{After: 1, First: 2},
			opts:            ListNotebooksOptions{},
			wantNotebookIDs: getNotebookIDs(createdNotebooks[1:3]),
			wantCount:       3,
		},
		{
			name:            "get notebooks page with options",
			userID:          user1.ID,
			pageOpts:        ListNotebooksPageOptions{After: 1, First: 1},
			opts:            ListNotebooksOptions{CreatorUserID: user1.ID},
			wantNotebookIDs: getNotebookIDs(createdNotebooks[1:2]),
			wantCount:       2,
		},
		{
			name:            "get user2 notebooks as user1",
			userID:          user1.ID,
			pageOpts:        ListNotebooksPageOptions{After: 0, First: 4},
			opts:            ListNotebooksOptions{CreatorUserID: user2.ID},
			wantNotebookIDs: getNotebookIDs(createdNotebooks[2:3]),
			wantCount:       1,
		},
		{
			name:            "query notebooks by title",
			userID:          user1.ID,
			pageOpts:        ListNotebooksPageOptions{After: 0, First: 4},
			opts:            ListNotebooksOptions{Query: "public"},
			wantNotebookIDs: []int64{createdNotebooks[0].ID, createdNotebooks[2].ID},
			wantCount:       2,
		},
		{
			name:            "query notebooks by title and creator user id",
			userID:          user1.ID,
			pageOpts:        ListNotebooksPageOptions{After: 0, First: 4},
			opts:            ListNotebooksOptions{Query: "public", CreatorUserID: user1.ID},
			wantNotebookIDs: []int64{createdNotebooks[0].ID},
			wantCount:       1,
		},
		{
			name:            "query notebook blocks by prefix",
			userID:          user2.ID,
			pageOpts:        ListNotebooksPageOptions{After: 0, First: 4},
			opts:            ListNotebooksOptions{Query: "lor"},
			wantNotebookIDs: []int64{createdNotebooks[0].ID},
			wantCount:       1,
		},
		{
			name:            "query notebook blocks case insensitively",
			userID:          user2.ID,
			pageOpts:        ListNotebooksPageOptions{After: 0, First: 4},
			opts:            ListNotebooksOptions{Query: "ADIPISC"},
			wantNotebookIDs: []int64{createdNotebooks[0].ID},
			wantCount:       1,
		},
		{
			name:            "query notebook blocks by multiple prefixes",
			userID:          user2.ID,
			pageOpts:        ListNotebooksPageOptions{After: 0, First: 4},
			opts:            ListNotebooksOptions{Query: "auc od"},
			wantNotebookIDs: []int64{createdNotebooks[2].ID},
			wantCount:       1,
		},
		{
			name:            "query notebook blocks by file path",
			userID:          user2.ID,
			pageOpts:        ListNotebooksPageOptions{After: 0, First: 4},
			opts:            ListNotebooksOptions{Query: "client/web/file.tsx"},
			wantNotebookIDs: getNotebookIDs(createdNotebooks[2:]),
			wantCount:       2,
		},
		{
			name:            "order by updated at ascending",
			userID:          user1.ID,
			pageOpts:        ListNotebooksPageOptions{First: 4},
			opts:            ListNotebooksOptions{OrderBy: NotebooksOrderByUpdatedAt, OrderByDescending: false},
			wantNotebookIDs: []int64{createdNotebooks[1].ID, createdNotebooks[0].ID, createdNotebooks[2].ID},
			wantCount:       3,
		},
		{
			name:            "order by updated at descending",
			userID:          user1.ID,
			pageOpts:        ListNotebooksPageOptions{First: 4},
			opts:            ListNotebooksOptions{OrderBy: NotebooksOrderByUpdatedAt, OrderByDescending: true},
			wantNotebookIDs: []int64{createdNotebooks[2].ID, createdNotebooks[0].ID, createdNotebooks[1].ID},
			wantCount:       3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: tt.userID})
			gotNotebooks, err := n.ListNotebooks(ctx, tt.pageOpts, tt.opts)
			if err != nil {
				t.Fatal(err)
			}
			gotNotebookIDs := getNotebookIDs(gotNotebooks)
			if !reflect.DeepEqual(tt.wantNotebookIDs, gotNotebookIDs) {
				t.Fatalf("wanted %+v ids, got %+v", tt.wantNotebookIDs, gotNotebookIDs)
			}
			gotNotebooksCount, err := n.CountNotebooks(ctx, tt.opts)
			if err != nil {
				t.Fatal(err)
			}
			if tt.wantCount != gotNotebooksCount {
				t.Fatalf("wanted %d notebooks, got %d", tt.wantCount, gotNotebooksCount)
			}
		})
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
		name       string
		notebookID int64
		userID     int32
		wantErr    *error
	}{
		{name: "user1 get user1 public notebook", notebookID: createdNotebooks[0].ID, userID: user1.ID, wantErr: nil},
		{name: "user1 get user1 private notebook", notebookID: createdNotebooks[1].ID, userID: user1.ID, wantErr: nil},
		// User2 *can* access a public notebook from a different user (User1)
		{name: "user2 get user1 public notebook", notebookID: createdNotebooks[0].ID, userID: user2.ID, wantErr: nil},
		// User2 *cannot* access a private notebook from a different user (User1)
		{name: "user2 get user1 private notebook", notebookID: createdNotebooks[1].ID, userID: user2.ID, wantErr: &ErrNotebookNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: tt.userID})
			_, err := n.GetNotebook(ctx, tt.notebookID)
			if tt.wantErr != nil && !errors.Is(err, *tt.wantErr) {
				t.Errorf("expected error not found in chain: got %+v, want %+v", err, *tt.wantErr)
			} else if tt.wantErr == nil && err != nil {
				t.Errorf("expected no error, got %+v", err)
			}
		})
	}
}

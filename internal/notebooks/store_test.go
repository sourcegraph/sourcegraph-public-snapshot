package notebooks

import (
	"context"
	"reflect"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

func notebookByUser(notebook *Notebook, userID int32) *Notebook {
	notebook.CreatorUserID = userID
	notebook.UpdaterUserID = userID
	notebook.NamespaceUserID = userID
	return notebook
}

func notebookByOrg(notebook *Notebook, creatorID int32, orgID int32) *Notebook {
	notebook.CreatorUserID = creatorID
	notebook.UpdaterUserID = creatorID
	notebook.NamespaceOrgID = orgID
	return notebook
}

func TestCreateAndGetNotebook(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())
	u := db.Users()
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
		{ID: "4", Type: NotebookSymbolBlockType, SymbolInput: &NotebookSymbolBlockInput{
			RepositoryName:      "github.com/sourcegraph/sourcegraph",
			FilePath:            "client/web/file.tsx",
			LineContext:         1,
			SymbolName:          "function",
			SymbolContainerName: "container",
			SymbolKind:          "FUNCTION",
		}},
	}
	notebook := notebookByUser(&Notebook{Title: "Notebook Title", Blocks: blocks, Public: true}, user.ID)
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
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())
	u := db.Users()
	n := Notebooks(db)

	user, err := u.Create(ctx, database.NewUser{Username: "u", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	blocks := NotebookBlocks{{ID: "1", Type: NotebookQueryBlockType, QueryInput: &NotebookQueryBlockInput{"repo:a b"}}}
	notebook := notebookByUser(&Notebook{Title: "Notebook Title", Blocks: blocks, Public: true}, user.ID)
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
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())
	u := db.Users()
	n := Notebooks(db)

	user, err := u.Create(ctx, database.NewUser{Username: "u", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	blocks := NotebookBlocks{{ID: "1", Type: NotebookQueryBlockType, QueryInput: &NotebookQueryBlockInput{"repo:a b"}}}
	notebook := notebookByUser(&Notebook{Title: "Notebook Title", Blocks: blocks, Public: true}, user.ID)
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

func createNotebookStars(ctx context.Context, store NotebooksStore, userID int32, notebookIDs ...int64) ([]*NotebookStar, error) {
	stars := make([]*NotebookStar, 0, len(notebookIDs))
	for _, id := range notebookIDs {
		star, err := store.CreateNotebookStar(ctx, id, userID)
		if err != nil {
			return nil, err
		}
		stars = append(stars, star)
	}
	return stars, nil
}

func TestListingAndCountingNotebooks(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	internalCtx := actor.WithInternalActor(context.Background())
	u := db.Users()
	o := db.Orgs()
	om := db.OrgMembers()
	n := Notebooks(db)

	user1, err := u.Create(internalCtx, database.NewUser{Username: "u1", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	user2, err := u.Create(internalCtx, database.NewUser{Username: "u2", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	displayName := "My Org"
	org, err := o.Create(internalCtx, "myorg", &displayName)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	_, err = om.Create(internalCtx, org.ID, user1.ID)
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
		notebookByUser(&Notebook{Title: "Notebook User1 Public", Blocks: NotebookBlocks{blocks[0], blocks[4]}, Public: true}, user1.ID),
		notebookByUser(&Notebook{Title: "Notebook User1 Private", Blocks: NotebookBlocks{blocks[1]}, Public: false}, user1.ID),
		notebookByUser(&Notebook{Title: "Notebook User2 Public", Blocks: NotebookBlocks{blocks[2], blocks[5]}, Public: true}, user2.ID),
		notebookByUser(&Notebook{Title: "Notebook User2 Private", Blocks: NotebookBlocks{blocks[3]}, Public: false}, user2.ID),
		notebookByOrg(&Notebook{Title: "Notebook Org Public", Blocks: NotebookBlocks{}, Public: true}, user1.ID, org.ID),
		notebookByOrg(&Notebook{Title: "Notebook Org Private", Blocks: NotebookBlocks{}, Public: false}, user1.ID, org.ID),
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = createNotebookStars(internalCtx, n, user1.ID, createdNotebooks[0].ID, createdNotebooks[2].ID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = createNotebookStars(internalCtx, n, user2.ID, createdNotebooks[2].ID)
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

	getNotebookIDs := func(indices ...int) []int64 {
		ids := make([]int64, 0, len(indices))
		for _, idx := range indices {
			ids = append(ids, createdNotebooks[idx].ID)
		}
		return ids
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
			wantNotebookIDs: getNotebookIDs(0, 1, 2, 4),
			wantCount:       5,
		},
		{
			// User2 should not have access to the private org notebook
			name:            "get all user2 accessible notebooks",
			userID:          user2.ID,
			pageOpts:        ListNotebooksPageOptions{First: 4},
			opts:            ListNotebooksOptions{},
			wantNotebookIDs: getNotebookIDs(0, 2, 3, 4),
			wantCount:       4,
		},
		{
			name:            "get notebooks page",
			userID:          user1.ID,
			pageOpts:        ListNotebooksPageOptions{After: 1, First: 2},
			opts:            ListNotebooksOptions{},
			wantNotebookIDs: getNotebookIDs(1, 2),
			wantCount:       5,
		},
		{
			name:            "get notebooks page with options",
			userID:          user1.ID,
			pageOpts:        ListNotebooksPageOptions{After: 1, First: 1},
			opts:            ListNotebooksOptions{CreatorUserID: user1.ID},
			wantNotebookIDs: getNotebookIDs(1),
			wantCount:       4,
		},
		{
			name:            "get user2 notebooks as user1",
			userID:          user1.ID,
			pageOpts:        ListNotebooksPageOptions{After: 0, First: 4},
			opts:            ListNotebooksOptions{CreatorUserID: user2.ID},
			wantNotebookIDs: getNotebookIDs(2),
			wantCount:       1,
		},
		{
			name:            "query notebooks by title",
			userID:          user1.ID,
			pageOpts:        ListNotebooksPageOptions{After: 0, First: 4},
			opts:            ListNotebooksOptions{Query: "public"},
			wantNotebookIDs: getNotebookIDs(0, 2, 4),
			wantCount:       3,
		},
		{
			name:            "query notebooks by title and creator user id",
			userID:          user1.ID,
			pageOpts:        ListNotebooksPageOptions{After: 0, First: 4},
			opts:            ListNotebooksOptions{Query: "public", CreatorUserID: user1.ID},
			wantNotebookIDs: getNotebookIDs(0, 4),
			wantCount:       2,
		},
		{
			name:            "query notebook blocks by prefix",
			userID:          user2.ID,
			pageOpts:        ListNotebooksPageOptions{After: 0, First: 4},
			opts:            ListNotebooksOptions{Query: "lor"},
			wantNotebookIDs: getNotebookIDs(0),
			wantCount:       1,
		},
		{
			name:            "query notebook blocks case insensitively",
			userID:          user2.ID,
			pageOpts:        ListNotebooksPageOptions{After: 0, First: 4},
			opts:            ListNotebooksOptions{Query: "ADIPISC"},
			wantNotebookIDs: getNotebookIDs(0),
			wantCount:       1,
		},
		{
			name:            "query notebook blocks by multiple prefixes",
			userID:          user2.ID,
			pageOpts:        ListNotebooksPageOptions{After: 0, First: 4},
			opts:            ListNotebooksOptions{Query: "auc od"},
			wantNotebookIDs: getNotebookIDs(2),
			wantCount:       1,
		},
		{
			name:            "query notebook blocks by file path",
			userID:          user2.ID,
			pageOpts:        ListNotebooksPageOptions{After: 0, First: 4},
			opts:            ListNotebooksOptions{Query: "client/web/file.tsx"},
			wantNotebookIDs: getNotebookIDs(2, 3),
			wantCount:       2,
		},
		{
			name:            "order by updated at ascending",
			userID:          user1.ID,
			pageOpts:        ListNotebooksPageOptions{First: 4},
			opts:            ListNotebooksOptions{OrderBy: NotebooksOrderByUpdatedAt, OrderByDescending: false},
			wantNotebookIDs: getNotebookIDs(1, 4, 5, 0),
			wantCount:       5,
		},
		{
			name:            "order by updated at descending",
			userID:          user1.ID,
			pageOpts:        ListNotebooksPageOptions{First: 4},
			opts:            ListNotebooksOptions{OrderBy: NotebooksOrderByUpdatedAt, OrderByDescending: true},
			wantNotebookIDs: getNotebookIDs(2, 0, 5, 4),
			wantCount:       5,
		},
		{
			name:            "order by notebook stars descending",
			userID:          user1.ID,
			pageOpts:        ListNotebooksPageOptions{First: 2},
			opts:            ListNotebooksOptions{OrderBy: NotebooksOrderByStarCount, OrderByDescending: true},
			wantNotebookIDs: getNotebookIDs(2, 0),
			wantCount:       5,
		},
		{
			name:            "filter notebooks if user has starred them",
			userID:          user1.ID,
			pageOpts:        ListNotebooksPageOptions{First: 4},
			opts:            ListNotebooksOptions{StarredByUserID: user1.ID},
			wantNotebookIDs: getNotebookIDs(0, 2),
			wantCount:       2,
		},
		{
			name:            "filter notebooks by user namespace",
			userID:          user1.ID,
			pageOpts:        ListNotebooksPageOptions{First: 2},
			opts:            ListNotebooksOptions{NamespaceUserID: user1.ID},
			wantNotebookIDs: getNotebookIDs(0, 1),
			wantCount:       2,
		},
		{
			name:            "user1 filter notebooks by org namespace",
			userID:          user1.ID,
			pageOpts:        ListNotebooksPageOptions{First: 2},
			opts:            ListNotebooksOptions{NamespaceOrgID: org.ID},
			wantNotebookIDs: getNotebookIDs(4, 5),
			wantCount:       2,
		},
		{
			// User2 is not a member of the org
			name:            "user2 filter notebooks by org namespace",
			userID:          user2.ID,
			pageOpts:        ListNotebooksPageOptions{First: 2},
			opts:            ListNotebooksOptions{NamespaceOrgID: org.ID},
			wantNotebookIDs: getNotebookIDs(4),
			wantCount:       1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: tt.userID})
			gotNotebooks, err := n.ListNotebooks(ctx, tt.pageOpts, tt.opts)
			if err != nil {
				t.Fatal(err)
			}
			gotNotebookIDs := make([]int64, 0, len(gotNotebooks))
			for _, notebook := range gotNotebooks {
				gotNotebookIDs = append(gotNotebookIDs, notebook.ID)
			}
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
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())
	u := db.Users()
	n := Notebooks(db)

	user, err := u.Create(ctx, database.NewUser{Username: "u", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	blocks := NotebookBlocks{{ID: "1", Type: NotebookQueryBlockType}}
	notebook := notebookByUser(&Notebook{Title: "Notebook Title", Blocks: blocks, Public: true}, user.ID)
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
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	internalCtx := actor.WithInternalActor(context.Background())
	u := db.Users()
	o := db.Orgs()
	om := db.OrgMembers()
	n := Notebooks(db)

	user1, err := u.Create(internalCtx, database.NewUser{Username: "u1", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	user2, err := u.Create(internalCtx, database.NewUser{Username: "u2", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	displayName := "My Org"
	org, err := o.Create(internalCtx, "myorg", &displayName)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	_, err = om.Create(internalCtx, org.ID, user1.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	createdNotebooks, err := createNotebooks(internalCtx, n, []*Notebook{
		notebookByUser(&Notebook{Title: "Notebook User1 Public", Blocks: NotebookBlocks{}, Public: true}, user1.ID),
		notebookByUser(&Notebook{Title: "Notebook User1 Private", Blocks: NotebookBlocks{}, Public: false}, user1.ID),
		notebookByOrg(&Notebook{Title: "Notebook User1 Org Public", Blocks: NotebookBlocks{}, Public: true}, user1.ID, org.ID),
		notebookByOrg(&Notebook{Title: "Notebook User1 Org Private", Blocks: NotebookBlocks{}, Public: false}, user1.ID, org.ID),
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
		{name: "user2 get org public notebook", notebookID: createdNotebooks[2].ID, userID: user2.ID, wantErr: nil},
		// User1 is a member of the org
		{name: "user1 get org private notebook", notebookID: createdNotebooks[3].ID, userID: user1.ID, wantErr: nil},
		// User2 is not a member of the org
		{name: "user2 get org private notebook", notebookID: createdNotebooks[3].ID, userID: user2.ID, wantErr: &ErrNotebookNotFound},
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

func TestListingNotebookStars(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	internalCtx := actor.WithInternalActor(context.Background())
	u := db.Users()
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
		notebookByUser(&Notebook{Title: "Notebook1", Blocks: NotebookBlocks{}, Public: true}, user1.ID),
		notebookByUser(&Notebook{Title: "Notebook2", Blocks: NotebookBlocks{}, Public: true}, user2.ID),
	})
	if err != nil {
		t.Fatal(err)
	}

	user1Stars, err := createNotebookStars(internalCtx, n, user1.ID, createdNotebooks[0].ID, createdNotebooks[1].ID)
	if err != nil {
		t.Fatal(err)
	}

	user2Stars, err := createNotebookStars(internalCtx, n, user2.ID, createdNotebooks[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		notebookID int64
		pageOpts   ListNotebookStarsPageOptions
		wantStars  []*NotebookStar
		wantCount  int64
	}{
		{
			name:       "get first notebook first stars page",
			notebookID: createdNotebooks[0].ID,
			pageOpts:   ListNotebookStarsPageOptions{First: 2},
			wantStars:  []*NotebookStar{user2Stars[0], user1Stars[0]},
			wantCount:  2,
		},
		{
			name:       "get first notebook second stars page",
			notebookID: createdNotebooks[0].ID,
			pageOpts:   ListNotebookStarsPageOptions{First: 1, After: 1},
			wantStars:  []*NotebookStar{user1Stars[0]},
			wantCount:  2,
		},
		{
			name:       "get second notebook first stars page",
			notebookID: createdNotebooks[1].ID,
			pageOpts:   ListNotebookStarsPageOptions{First: 1},
			wantStars:  []*NotebookStar{user1Stars[1]},
			wantCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStars, err := n.ListNotebookStars(internalCtx, tt.pageOpts, tt.notebookID)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(tt.wantStars, gotStars) {
				t.Fatalf("wanted %+v stars, got %+v", tt.wantStars, gotStars)
			}

			gotCountStars, err := n.CountNotebookStars(internalCtx, tt.notebookID)
			if err != nil {
				t.Fatal(err)
			}
			if tt.wantCount != gotCountStars {
				t.Fatalf("wanted %d stars count, got %d", tt.wantCount, gotCountStars)
			}
		})
	}
}

func TestCreatingAndDeletingNotebookStars(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	internalCtx := actor.WithInternalActor(context.Background())
	u := db.Users()
	n := Notebooks(db)

	user, err := u.Create(internalCtx, database.NewUser{Username: "u1", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	createdNotebooks, err := createNotebooks(internalCtx, n, []*Notebook{
		notebookByUser(&Notebook{Title: "Notebook", Blocks: NotebookBlocks{}, Public: true}, user.ID),
		notebookByUser(&Notebook{Title: "Notebook", Blocks: NotebookBlocks{}, Public: true}, user.ID),
	})
	if err != nil {
		t.Fatal(err)
	}
	// Use the second notebook, so the user.ID and notebook.ID are different.
	notebook := createdNotebooks[1]

	_, err = n.CreateNotebookStar(internalCtx, notebook.ID, user.ID)
	if err != nil {
		t.Fatal(err)
	}

	// User cannot create multiple stars for the same notebook
	_, err = n.CreateNotebookStar(internalCtx, notebook.ID, user.ID)
	if err == nil {
		t.Errorf("expected non-nil error, got nil")
	}

	_, err = n.GetNotebookStar(internalCtx, notebook.ID, user.ID)
	if err != nil {
		t.Errorf("expected to get notebook star, got %+v", err)
	}

	err = n.DeleteNotebookStar(internalCtx, notebook.ID, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = n.GetNotebookStar(internalCtx, notebook.ID, user.ID)
	if err == nil {
		t.Errorf("expected non-nil error, got nil")
	}
}

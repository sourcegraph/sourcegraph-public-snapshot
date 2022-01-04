package resolvers

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/resolvers/apitest"
	notebooksapitest "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/notebooks/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/notebooks"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

const notebookFields = `
	id
	title
	creator {
		username
	}
	createdAt
	updatedAt
	public
	viewerCanManage
	blocks {
		... on MarkdownBlock {
			__typename
			id
			markdownInput
		}
		... on QueryBlock {
			__typename
			id
			queryInput
		}
		... on FileBlock {
			__typename
			id
			fileInput {
				repositoryName
				filePath
				revision
				lineRange {
					startLine
					endLine
				}
			}
		}
	}
`

var queryNotebook = fmt.Sprintf(`
query Notebook($id: ID!) {
	node(id: $id) {
		... on Notebook {
			%s
		}
	}
}
`, notebookFields)

var listNotebooksQuery = fmt.Sprintf(`
query Notebooks($first: Int!, $after: String, $orderBy: NotebooksOrderBy, $descending: Boolean, $creatorUserID: ID, $query: String) {
	notebooks(first: $first, after: $after, orderBy: $orderBy, descending: $descending, creatorUserID: $creatorUserID, query: $query) {
		nodes {
			%s
	  	}
	  	totalCount
		pageInfo {
			endCursor
			hasNextPage
	  	}
	}
}
`, notebookFields)

var createNotebookMutation = fmt.Sprintf(`
mutation CreateNotebook($notebook: NotebookInput!) {
	createNotebook(notebook: $notebook) {
		%s
	}
}
`, notebookFields)

var updateNotebookMutation = fmt.Sprintf(`
mutation UpdateNotebook($id: ID!, $notebook: NotebookInput!) {
	updateNotebook(id: $id, notebook: $notebook) {
		%s
	}
}
`, notebookFields)

const deleteNotebookMutation = `
mutation DeleteNotebook($id: ID!) {
	deleteNotebook(id: $id) {
		alwaysNil
	}
}
`

func notebookFixture(userID int32, public bool) *notebooks.Notebook {
	revision := "deadbeef"
	blocks := notebooks.NotebookBlocks{
		{ID: "1", Type: notebooks.NotebookQueryBlockType, QueryInput: &notebooks.NotebookQueryBlockInput{Text: "repo:a b"}},
		{ID: "2", Type: notebooks.NotebookMarkdownBlockType, MarkdownInput: &notebooks.NotebookMarkdownBlockInput{Text: "# Title"}},
		{ID: "3", Type: notebooks.NotebookFileBlockType, FileInput: &notebooks.NotebookFileBlockInput{
			RepositoryName: "github.com/sourcegraph/sourcegraph",
			FilePath:       "client/web/file.tsx",
			Revision:       &revision,
			LineRange:      &notebooks.LineRange{StartLine: 10, EndLine: 12},
		}},
	}
	return &notebooks.Notebook{Title: "Notebook Title", Blocks: blocks, Public: public, CreatorUserID: userID}
}

func compareNotebookAPIResponses(t *testing.T, wantNotebookResponse notebooksapitest.Notebook, gotNotebookResponse notebooksapitest.Notebook, ignoreIDAndTimestamps bool) {
	t.Helper()
	if ignoreIDAndTimestamps {
		// Ignore ID and timestamps for easier comparison
		wantNotebookResponse.ID = gotNotebookResponse.ID
		wantNotebookResponse.CreatedAt = gotNotebookResponse.CreatedAt
		wantNotebookResponse.UpdatedAt = gotNotebookResponse.UpdatedAt
	}

	if diff := cmp.Diff(wantNotebookResponse, gotNotebookResponse); diff != "" {
		t.Fatalf("wrong notebook response (-want +got):\n%s", diff)
	}
}

func TestGetNotebook(t *testing.T) {
	db := dbtest.NewDB(t)
	ctx := actor.WithInternalActor(context.Background())
	u := database.Users(db)
	n := notebooks.Notebooks(db)

	user, err := u.Create(ctx, database.NewUser{Username: "u", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	createdNotebook, err := n.CreateNotebook(ctx, notebookFixture(user.ID, true))
	if err != nil {
		t.Fatal(err)
	}

	database := database.NewDB(db)
	schema, err := graphqlbackend.NewSchema(database, nil, nil, nil, nil, nil, nil, nil, nil, nil, NewResolver(database))
	if err != nil {
		t.Fatal(err)
	}

	notebookGQLID := marshalNotebookID(createdNotebook.ID)
	input := map[string]interface{}{"id": notebookGQLID}
	var response struct{ Node notebooksapitest.Notebook }
	apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(user.ID)), t, schema, input, &response, queryNotebook)

	wantNotebookResponse := notebooksapitest.NotebookToAPIResponse(createdNotebook, notebookGQLID, user.Username, true)
	compareNotebookAPIResponses(t, wantNotebookResponse, response.Node, false)
}

func TestCreateNotebook(t *testing.T) {
	db := dbtest.NewDB(t)
	ctx := actor.WithInternalActor(context.Background())
	u := database.Users(db)

	user, err := u.Create(ctx, database.NewUser{Username: "u", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	database := database.NewDB(db)
	schema, err := graphqlbackend.NewSchema(database, nil, nil, nil, nil, nil, nil, nil, nil, nil, NewResolver(database))
	if err != nil {
		t.Fatal(err)
	}

	notebook := notebookFixture(user.ID, true)
	input := map[string]interface{}{"notebook": notebooksapitest.NotebookToAPIInput(notebook)}
	var response struct{ CreateNotebook notebooksapitest.Notebook }
	apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(user.ID)), t, schema, input, &response, createNotebookMutation)

	wantNotebookResponse := notebooksapitest.NotebookToAPIResponse(notebook, marshalNotebookID(notebook.ID), user.Username, true)
	compareNotebookAPIResponses(t, wantNotebookResponse, response.CreateNotebook, true)
}

func TestUpdateNotebook(t *testing.T) {
	db := dbtest.NewDB(t)
	internalCtx := actor.WithInternalActor(context.Background())
	u := database.Users(db)
	n := notebooks.Notebooks(db)

	user1, err := u.Create(internalCtx, database.NewUser{Username: "u1", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	user2, err := u.Create(internalCtx, database.NewUser{Username: "u2", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	database := database.NewDB(db)
	schema, err := graphqlbackend.NewSchema(database, nil, nil, nil, nil, nil, nil, nil, nil, nil, NewResolver(database))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name            string
		publicNotebook  bool
		creatorID       int32
		creatorUsername string
		updaterID       int32
		wantErr         string
	}{
		{
			name:            "user can update their own public notebook",
			publicNotebook:  true,
			creatorID:       user1.ID,
			creatorUsername: user1.Username,
			updaterID:       user1.ID,
		},
		{
			name:            "user can update their own private notebook",
			publicNotebook:  false,
			creatorID:       user1.ID,
			creatorUsername: user1.Username,
			updaterID:       user1.ID,
		},
		{
			name:            "user cannot update other public notebooks",
			publicNotebook:  true,
			creatorID:       user1.ID,
			creatorUsername: user1.Username,
			updaterID:       user2.ID,
			wantErr:         "user does not have permissions to update the notebook",
		},
		{
			name:            "user cannot update other private notebooks",
			publicNotebook:  false,
			creatorID:       user1.ID,
			creatorUsername: user1.Username,
			updaterID:       user2.ID,
			wantErr:         "notebook not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdNotebook, err := n.CreateNotebook(internalCtx, notebookFixture(tt.creatorID, tt.publicNotebook))
			if err != nil {
				t.Fatal(err)
			}

			updatedNotebook := createdNotebook
			updatedNotebook.Title = "Updated Title"
			updatedNotebook.Public = !createdNotebook.Public
			updatedNotebook.Blocks = createdNotebook.Blocks[:1]

			input := map[string]interface{}{"id": marshalNotebookID(createdNotebook.ID), "notebook": notebooksapitest.NotebookToAPIInput(updatedNotebook)}
			var response struct{ UpdateNotebook notebooksapitest.Notebook }
			gotErrors := apitest.Exec(actor.WithActor(context.Background(), actor.FromUser(tt.updaterID)), t, schema, input, &response, updateNotebookMutation)

			if tt.wantErr != "" && len(gotErrors) == 0 {
				t.Fatal("expected error, got none")
			}

			if tt.wantErr != "" && !strings.Contains(gotErrors[0].Message, tt.wantErr) {
				t.Fatalf("expected error containing '%s', got '%s'", tt.wantErr, gotErrors[0].Message)
			}

			if tt.wantErr == "" {
				wantNotebookResponse := notebooksapitest.NotebookToAPIResponse(updatedNotebook, marshalNotebookID(updatedNotebook.ID), tt.creatorUsername, tt.creatorID == tt.updaterID)
				compareNotebookAPIResponses(t, wantNotebookResponse, response.UpdateNotebook, true)
			}
		})
	}
}

func TestDeleteNotebook(t *testing.T) {
	db := dbtest.NewDB(t)
	internalCtx := actor.WithInternalActor(context.Background())
	u := database.Users(db)
	n := notebooks.Notebooks(db)

	user1, err := u.Create(internalCtx, database.NewUser{Username: "u1", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	user2, err := u.Create(internalCtx, database.NewUser{Username: "u2", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	database := database.NewDB(db)
	schema, err := graphqlbackend.NewSchema(database, nil, nil, nil, nil, nil, nil, nil, nil, nil, NewResolver(database))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name           string
		publicNotebook bool
		creatorID      int32
		deleterID      int32
		wantErr        string
	}{
		{
			name:           "user can delete their own public notebook",
			publicNotebook: true,
			creatorID:      user1.ID,
			deleterID:      user1.ID,
		},
		{
			name:           "user can delete their own private notebook",
			publicNotebook: false,
			creatorID:      user1.ID,
			deleterID:      user1.ID,
		},
		{
			name:           "user cannot delete other public notebooks",
			publicNotebook: true,
			creatorID:      user1.ID,
			deleterID:      user2.ID,
			wantErr:        "user does not have permissions to update the notebook",
		},
		{
			name:           "user cannot delete other private notebooks",
			publicNotebook: false,
			creatorID:      user1.ID,
			deleterID:      user2.ID,
			wantErr:        "notebook not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdNotebook, err := n.CreateNotebook(internalCtx, notebookFixture(tt.creatorID, tt.publicNotebook))
			if err != nil {
				t.Fatal(err)
			}

			input := map[string]interface{}{"id": marshalNotebookID(createdNotebook.ID)}
			var response struct{}
			gotErrors := apitest.Exec(actor.WithActor(context.Background(), actor.FromUser(tt.deleterID)), t, schema, input, &response, deleteNotebookMutation)

			if tt.wantErr != "" && len(gotErrors) == 0 {
				t.Fatal("expected error, got none")
			}

			if tt.wantErr != "" && !strings.Contains(gotErrors[0].Message, tt.wantErr) {
				t.Fatalf("expected error containing '%s', got '%s'", tt.wantErr, gotErrors[0].Message)
			}

			_, err = n.GetNotebook(actor.WithActor(context.Background(), actor.FromUser(tt.creatorID)), createdNotebook.ID)
			if tt.wantErr == "" && !errors.Is(err, notebooks.ErrNotebookNotFound) {
				t.Fatal("expected to not find a deleted notebook")
			}
		})
	}
}

func createNotebooks(t *testing.T, db *sql.DB, notebooksToCreate []*notebooks.Notebook) []*notebooks.Notebook {
	t.Helper()
	n := notebooks.Notebooks(db)
	internalCtx := actor.WithInternalActor(context.Background())
	createdNotebooks := make([]*notebooks.Notebook, 0, len(notebooksToCreate))
	for _, notebook := range notebooksToCreate {
		createdNotebook, err := n.CreateNotebook(internalCtx, notebook)
		if err != nil {
			t.Fatal(err)
		}
		createdNotebooks = append(createdNotebooks, createdNotebook)
	}
	return createdNotebooks
}

func TestListNotebooks(t *testing.T) {
	db := dbtest.NewDB(t)
	internalCtx := actor.WithInternalActor(context.Background())
	u := database.Users(db)

	user1, err := u.Create(internalCtx, database.NewUser{Username: "u1", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	user2, err := u.Create(internalCtx, database.NewUser{Username: "u2", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	idToUsername := map[int32]string{user1.ID: user1.Username, user2.ID: user2.Username}

	n1 := notebookFixture(user1.ID, true)
	n1.Blocks = notebooks.NotebookBlocks{{ID: "1", Type: notebooks.NotebookMarkdownBlockType, MarkdownInput: &notebooks.NotebookMarkdownBlockInput{Text: "# A special title"}}}

	createdNotebooks := createNotebooks(t, db, []*notebooks.Notebook{
		n1,
		notebookFixture(user1.ID, false),
		notebookFixture(user2.ID, true),
	})

	database := database.NewDB(db)
	schema, err := graphqlbackend.NewSchema(database, nil, nil, nil, nil, nil, nil, nil, nil, nil, NewResolver(database))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		viewerID      int32
		args          map[string]interface{}
		wantCount     int32
		wantNotebooks []*notebooks.Notebook
	}{
		{
			name:          "list all available notebooks",
			viewerID:      user1.ID,
			args:          map[string]interface{}{"first": 3, "orderBy": graphqlbackend.NotebookOrderByCreatedAt, "descending": false},
			wantNotebooks: createdNotebooks,
			wantCount:     3,
		},
		{
			name:          "list second page of available notebooks",
			viewerID:      user1.ID,
			args:          map[string]interface{}{"first": 2, "after": marshalNotebookCursor(1), "orderBy": graphqlbackend.NotebookOrderByCreatedAt, "descending": false},
			wantNotebooks: createdNotebooks[1:],
			wantCount:     3,
		},
		{
			name:          "query by block contents",
			viewerID:      user1.ID,
			args:          map[string]interface{}{"first": 3, "query": "special", "orderBy": graphqlbackend.NotebookOrderByCreatedAt, "descending": false},
			wantNotebooks: createdNotebooks[:1],
			wantCount:     1,
		},
		{
			name:          "filter by creator",
			viewerID:      user1.ID,
			args:          map[string]interface{}{"first": 3, "creatorUserID": graphqlbackend.MarshalUserID(user2.ID), "orderBy": graphqlbackend.NotebookOrderByCreatedAt, "descending": false},
			wantNotebooks: createdNotebooks[2:],
			wantCount:     1,
		},
		{
			name:          "user2 cannot view user1 private notebooks",
			viewerID:      user2.ID,
			args:          map[string]interface{}{"first": 3, "orderBy": graphqlbackend.NotebookOrderByCreatedAt, "descending": false},
			wantNotebooks: []*notebooks.Notebook{createdNotebooks[0], createdNotebooks[2]},
			wantCount:     2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response struct {
				Notebooks struct {
					Nodes      []notebooksapitest.Notebook
					TotalCount int32
					PageInfo   apitest.PageInfo
				}
			}
			apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(tt.viewerID)), t, schema, tt.args, &response, listNotebooksQuery)

			if len(tt.wantNotebooks) != len(response.Notebooks.Nodes) {
				t.Fatalf("wanted %d notebook nodes, got %d", len(tt.wantNotebooks), len(response.Notebooks.Nodes))
			}

			if tt.wantCount != response.Notebooks.TotalCount {
				t.Fatalf("wanted %d notebook total count, got %d", tt.wantCount, response.Notebooks.TotalCount)
			}

			for idx, createdNotebook := range tt.wantNotebooks {
				wantNotebookResponse := notebooksapitest.NotebookToAPIResponse(
					createdNotebook,
					marshalNotebookID(createdNotebook.ID),
					idToUsername[createdNotebook.CreatorUserID],
					createdNotebook.CreatorUserID == tt.viewerID,
				)
				compareNotebookAPIResponses(t, wantNotebookResponse, response.Notebooks.Nodes[idx], true)
			}
		})
	}
}

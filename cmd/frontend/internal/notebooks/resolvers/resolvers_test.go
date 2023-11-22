package resolvers

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/batches/resolvers/apitest"
	notebooksapitest "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/notebooks/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/notebooks"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const notebookFields = `
	id
	title
	creator {
		username
	}
	updater {
		username
	}
	createdAt
	updatedAt
	public
	viewerCanManage
	viewerHasStarred
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
		... on SymbolBlock {
			__typename
			id
			symbolInput {
				repositoryName
				filePath
				revision
				lineContext
				symbolName
				symbolContainerName
				symbolKind
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
query Notebooks($first: Int!, $after: String, $orderBy: NotebooksOrderBy, $descending: Boolean, $starredByUserID: ID, $creatorUserID: ID, $namespace: ID, $query: String) {
	notebooks(first: $first, after: $after, orderBy: $orderBy, descending: $descending, starredByUserID: $starredByUserID, creatorUserID: $creatorUserID, namespace: $namespace, query: $query) {
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

func notebookFixture(creatorID int32, namespaceUserID int32, namespaceOrgID int32, public bool) *notebooks.Notebook {
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
		{ID: "4", Type: notebooks.NotebookSymbolBlockType, SymbolInput: &notebooks.NotebookSymbolBlockInput{
			RepositoryName:      "github.com/sourcegraph/sourcegraph",
			FilePath:            "client/web/file.tsx",
			Revision:            &revision,
			LineContext:         1,
			SymbolName:          "function",
			SymbolContainerName: "container",
			SymbolKind:          "FUNCTION",
		}},
	}
	return &notebooks.Notebook{Title: "Notebook Title", Blocks: blocks, Public: public, CreatorUserID: creatorID, UpdaterUserID: creatorID, NamespaceUserID: namespaceUserID, NamespaceOrgID: namespaceOrgID}
}

func userNotebookFixture(userID int32, public bool) *notebooks.Notebook {
	return notebookFixture(userID, userID, 0, public)
}

func orgNotebookFixture(creatorID int32, orgID int32, public bool) *notebooks.Notebook {
	return notebookFixture(creatorID, 0, orgID, public)
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

func TestSingleNotebookCRUD(t *testing.T) {
	logger := logtest.Scoped(t)
	internalCtx := actor.WithInternalActor(context.Background())
	db := database.NewDB(logger, dbtest.NewDB(t))
	u := db.Users()
	o := db.Orgs()
	om := db.OrgMembers()

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

	schema, err := graphqlbackend.NewSchemaWithNotebooksResolver(db, NewResolver(db))
	if err != nil {
		t.Fatal(err)
	}

	testGetNotebook(t, db, schema, user1)
	testCreateNotebook(t, schema, user1, user2, org)
	testUpdateNotebook(t, db, schema, user1, user2, org)
	testDeleteNotebook(t, db, schema, user1, user2, org)
}

func testGetNotebook(t *testing.T, db database.DB, schema *graphql.Schema, user *types.User) {
	ctx := actor.WithInternalActor(context.Background())
	n := notebooks.Notebooks(db)

	createdNotebook, err := n.CreateNotebook(ctx, userNotebookFixture(user.ID, true))
	if err != nil {
		t.Fatal(err)
	}

	notebookGQLID := marshalNotebookID(createdNotebook.ID)
	input := map[string]any{"id": notebookGQLID}
	var response struct{ Node notebooksapitest.Notebook }
	apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(user.ID)), t, schema, input, &response, queryNotebook)

	wantNotebookResponse := notebooksapitest.NotebookToAPIResponse(createdNotebook, notebookGQLID, user.Username, user.Username, true)
	compareNotebookAPIResponses(t, wantNotebookResponse, response.Node, false)
}

func testCreateNotebook(t *testing.T, schema *graphql.Schema, user1 *types.User, user2 *types.User, org *types.Org) {
	tests := []struct {
		name            string
		namespaceUserID int32
		namespaceOrgID  int32
		creator         *types.User
		wantErr         string
	}{
		{
			name:            "user can create a notebook in their namespace",
			namespaceUserID: user1.ID,
			creator:         user1,
		},
		{
			name:           "user can create a notebook in org namespace",
			namespaceOrgID: org.ID,
			creator:        user1,
		},
		{
			name:            "user2 cannot create a notebook in user1 namespace",
			namespaceUserID: user1.ID,
			creator:         user2,
			wantErr:         "user does not match the notebook user namespace",
		},
		{
			name:           "user2 cannot create a notebook in org namespace",
			namespaceOrgID: org.ID,
			creator:        user2,
			wantErr:        "user is not a member of the notebook organization namespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notebook := notebookFixture(tt.creator.ID, tt.namespaceUserID, tt.namespaceOrgID, true)
			input := map[string]any{"notebook": notebooksapitest.NotebookToAPIInput(notebook)}
			var response struct{ CreateNotebook notebooksapitest.Notebook }
			gotErrors := apitest.Exec(actor.WithActor(context.Background(), actor.FromUser(tt.creator.ID)), t, schema, input, &response, createNotebookMutation)

			if tt.wantErr != "" && len(gotErrors) == 0 {
				t.Fatal("expected error, got none")
			}

			if tt.wantErr != "" && !strings.Contains(gotErrors[0].Message, tt.wantErr) {
				t.Fatalf("expected error containing '%s', got '%s'", tt.wantErr, gotErrors[0].Message)
			}

			if tt.wantErr == "" {
				wantNotebookResponse := notebooksapitest.NotebookToAPIResponse(notebook, marshalNotebookID(notebook.ID), tt.creator.Username, tt.creator.Username, true)
				compareNotebookAPIResponses(t, wantNotebookResponse, response.CreateNotebook, true)
			}
		})
	}
}

func testUpdateNotebook(t *testing.T, db database.DB, schema *graphql.Schema, user1 *types.User, user2 *types.User, org *types.Org) {
	internalCtx := actor.WithInternalActor(context.Background())
	n := notebooks.Notebooks(db)

	tests := []struct {
		name                   string
		publicNotebook         bool
		creator                *types.User
		updater                *types.User
		namespaceUserID        int32
		namespaceOrgID         int32
		updatedNamespaceUserID int32
		updatedNamespaceOrgID  int32
		wantErr                string
	}{
		{
			name:            "user can update their own public notebook",
			publicNotebook:  true,
			creator:         user1,
			updater:         user1,
			namespaceUserID: user1.ID,
		},
		{
			name:            "user can update their own private notebook",
			publicNotebook:  false,
			creator:         user1,
			updater:         user1,
			namespaceUserID: user1.ID,
		},
		{
			name:           "user1 can update org public notebook",
			publicNotebook: true,
			creator:        user1,
			updater:        user1,
			namespaceOrgID: org.ID,
		},
		{
			name:           "user1 can update org private notebook",
			publicNotebook: false,
			creator:        user1,
			updater:        user1,
			namespaceOrgID: org.ID,
		},
		{
			name:            "user cannot update other users public notebooks",
			publicNotebook:  true,
			creator:         user1,
			updater:         user2,
			namespaceUserID: user1.ID,
			wantErr:         "user does not match the notebook user namespace",
		},
		{
			name:            "user cannot update other users private notebooks",
			publicNotebook:  false,
			creator:         user1,
			updater:         user2,
			namespaceUserID: user1.ID,
			wantErr:         "notebook not found",
		},
		{
			name:           "user2 cannot update org public notebook",
			publicNotebook: true,
			creator:        user1,
			updater:        user2,
			namespaceOrgID: org.ID,
			wantErr:        "user is not a member of the notebook organization namespace",
		},
		{
			name:           "user2 cannot update org private notebook",
			publicNotebook: false,
			creator:        user1,
			updater:        user2,
			namespaceOrgID: org.ID,
			wantErr:        "notebook not found",
		},
		{
			name:                  "change notebook user namespace to org namespace",
			publicNotebook:        true,
			creator:               user1,
			updater:               user1,
			namespaceUserID:       user1.ID,
			updatedNamespaceOrgID: org.ID,
		},
		{
			name:                   "change notebook org namespace to user namespace",
			publicNotebook:         true,
			creator:                user1,
			updater:                user1,
			namespaceOrgID:         org.ID,
			updatedNamespaceUserID: user1.ID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdNotebook, err := n.CreateNotebook(internalCtx, notebookFixture(tt.creator.ID, tt.namespaceUserID, tt.namespaceOrgID, tt.publicNotebook))
			if err != nil {
				t.Fatal(err)
			}

			updatedNotebook := createdNotebook
			updatedNotebook.Title = "Updated Title"
			updatedNotebook.Public = !createdNotebook.Public
			updatedNotebook.Blocks = createdNotebook.Blocks[:1]
			if tt.updatedNamespaceUserID != 0 || tt.updatedNamespaceOrgID != 0 {
				updatedNotebook.NamespaceUserID = tt.updatedNamespaceUserID
				updatedNotebook.NamespaceOrgID = tt.updatedNamespaceOrgID
			}

			input := map[string]any{"id": marshalNotebookID(createdNotebook.ID), "notebook": notebooksapitest.NotebookToAPIInput(updatedNotebook)}
			var response struct{ UpdateNotebook notebooksapitest.Notebook }
			gotErrors := apitest.Exec(actor.WithActor(context.Background(), actor.FromUser(tt.updater.ID)), t, schema, input, &response, updateNotebookMutation)

			if tt.wantErr != "" && len(gotErrors) == 0 {
				t.Fatal("expected error, got none")
			}

			if tt.wantErr != "" && !strings.Contains(gotErrors[0].Message, tt.wantErr) {
				t.Fatalf("expected error containing '%s', got '%s'", tt.wantErr, gotErrors[0].Message)
			}

			if tt.wantErr == "" {
				wantNotebookResponse := notebooksapitest.NotebookToAPIResponse(updatedNotebook, marshalNotebookID(updatedNotebook.ID), tt.creator.Username, tt.updater.Username, tt.creator.ID == tt.updater.ID)
				compareNotebookAPIResponses(t, wantNotebookResponse, response.UpdateNotebook, true)
			}
		})
	}
}

func testDeleteNotebook(t *testing.T, db database.DB, schema *graphql.Schema, user1 *types.User, user2 *types.User, org *types.Org) {
	internalCtx := actor.WithInternalActor(context.Background())
	n := notebooks.Notebooks(db)

	tests := []struct {
		name            string
		publicNotebook  bool
		creatorID       int32
		namespaceUserID int32
		namespaceOrgID  int32
		deleterID       int32
		wantErr         string
	}{
		{
			name:            "user can delete their own public notebook",
			publicNotebook:  true,
			creatorID:       user1.ID,
			namespaceUserID: user1.ID,
			deleterID:       user1.ID,
		},
		{
			name:            "user can delete their own private notebook",
			publicNotebook:  false,
			creatorID:       user1.ID,
			namespaceUserID: user1.ID,
			deleterID:       user1.ID,
		},
		{
			name:           "user1 can delete org public notebook",
			publicNotebook: true,
			creatorID:      user1.ID,
			namespaceOrgID: org.ID,
			deleterID:      user1.ID,
		},
		{
			name:           "user1 can delete org private notebook",
			publicNotebook: false,
			creatorID:      user1.ID,
			namespaceOrgID: org.ID,
			deleterID:      user1.ID,
		},
		{
			name:            "user2 cannot delete other user1 public notebook",
			publicNotebook:  true,
			creatorID:       user1.ID,
			namespaceUserID: user1.ID,
			deleterID:       user2.ID,
			wantErr:         "user does not match the notebook user namespace",
		},
		{
			name:            "user2 cannot delete other user1 private notebook",
			publicNotebook:  false,
			creatorID:       user1.ID,
			namespaceUserID: user1.ID,
			deleterID:       user2.ID,
			wantErr:         "notebook not found",
		},
		{
			name:           "user2 cannot delete org public notebook",
			publicNotebook: true,
			creatorID:      user1.ID,
			namespaceOrgID: org.ID,
			deleterID:      user2.ID,
			wantErr:        "user is not a member of the notebook organization namespace",
		},
		{
			name:           "user2 cannot delete org private notebook",
			publicNotebook: false,
			creatorID:      user1.ID,
			namespaceOrgID: org.ID,
			deleterID:      user2.ID,
			wantErr:        "notebook not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdNotebook, err := n.CreateNotebook(internalCtx, notebookFixture(tt.creatorID, tt.namespaceUserID, tt.namespaceOrgID, tt.publicNotebook))
			if err != nil {
				t.Fatal(err)
			}

			input := map[string]any{"id": marshalNotebookID(createdNotebook.ID)}
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

func createNotebooks(t *testing.T, db database.DB, notebooksToCreate []*notebooks.Notebook) []*notebooks.Notebook {
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

func createNotebookStars(t *testing.T, db database.DB, notebookID int64, userIDs ...int32) {
	t.Helper()
	n := notebooks.Notebooks(db)
	internalCtx := actor.WithInternalActor(context.Background())
	for _, userID := range userIDs {
		_, err := n.CreateNotebookStar(internalCtx, notebookID, userID)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestListNotebooks(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	internalCtx := actor.WithInternalActor(context.Background())
	u := db.Users()
	o := db.Orgs()
	om := db.OrgMembers()

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

	idToUsername := map[int32]string{user1.ID: user1.Username, user2.ID: user2.Username}

	n1 := userNotebookFixture(user1.ID, true)
	n1.Blocks = notebooks.NotebookBlocks{{ID: "1", Type: notebooks.NotebookMarkdownBlockType, MarkdownInput: &notebooks.NotebookMarkdownBlockInput{Text: "# A special title"}}}

	createdNotebooks := createNotebooks(t, db, []*notebooks.Notebook{
		n1,
		userNotebookFixture(user1.ID, false),
		userNotebookFixture(user2.ID, true),
		orgNotebookFixture(user1.ID, org.ID, false),
		orgNotebookFixture(user1.ID, org.ID, true),
	})
	createNotebookStars(t, db, createdNotebooks[0].ID, user1.ID)
	createNotebookStars(t, db, createdNotebooks[2].ID, user1.ID, user2.ID)

	getNotebooks := func(indices ...int) []*notebooks.Notebook {
		ids := make([]*notebooks.Notebook, 0, len(indices))
		for _, idx := range indices {
			ids = append(ids, createdNotebooks[idx])
		}
		return ids
	}

	schema, err := graphqlbackend.NewSchemaWithNotebooksResolver(db, NewResolver(db))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		viewerID      int32
		args          map[string]any
		wantCount     int32
		wantNotebooks []*notebooks.Notebook
	}{
		{
			name:          "list all available notebooks",
			viewerID:      user1.ID,
			args:          map[string]any{"first": 3, "orderBy": graphqlbackend.NotebookOrderByCreatedAt, "descending": false},
			wantNotebooks: getNotebooks(0, 1, 2),
			wantCount:     5,
		},
		{
			name:          "list second page of available notebooks",
			viewerID:      user1.ID,
			args:          map[string]any{"first": 2, "after": marshalNotebookCursor(1), "orderBy": graphqlbackend.NotebookOrderByCreatedAt, "descending": false},
			wantNotebooks: getNotebooks(1, 2),
			wantCount:     5,
		},
		{
			name:          "query by block contents",
			viewerID:      user1.ID,
			args:          map[string]any{"first": 3, "query": "special", "orderBy": graphqlbackend.NotebookOrderByCreatedAt, "descending": false},
			wantNotebooks: getNotebooks(0),
			wantCount:     1,
		},
		{
			name:          "filter by creator",
			viewerID:      user1.ID,
			args:          map[string]any{"first": 3, "creatorUserID": graphqlbackend.MarshalUserID(user2.ID), "orderBy": graphqlbackend.NotebookOrderByCreatedAt, "descending": false},
			wantNotebooks: getNotebooks(2),
			wantCount:     1,
		},
		{
			name:          "filter by user namespace",
			viewerID:      user1.ID,
			args:          map[string]any{"first": 3, "namespace": graphqlbackend.MarshalUserID(user1.ID), "orderBy": graphqlbackend.NotebookOrderByCreatedAt, "descending": false},
			wantNotebooks: getNotebooks(0, 1),
			wantCount:     2,
		},
		{
			name:          "filter by org namespace",
			viewerID:      user1.ID,
			args:          map[string]any{"first": 3, "namespace": graphqlbackend.MarshalOrgID(org.ID), "orderBy": graphqlbackend.NotebookOrderByCreatedAt, "descending": false},
			wantNotebooks: getNotebooks(3, 4),
			wantCount:     2,
		},
		{
			name:          "user2 cannot view user1 private notebooks",
			viewerID:      user2.ID,
			args:          map[string]any{"first": 3, "namespace": graphqlbackend.MarshalUserID(user1.ID), "orderBy": graphqlbackend.NotebookOrderByCreatedAt, "descending": false},
			wantNotebooks: getNotebooks(0),
			wantCount:     1,
		},
		{
			name:          "user2 cannot view org private notebooks",
			viewerID:      user2.ID,
			args:          map[string]any{"first": 3, "namespace": graphqlbackend.MarshalOrgID(org.ID), "orderBy": graphqlbackend.NotebookOrderByCreatedAt, "descending": false},
			wantNotebooks: getNotebooks(4),
			wantCount:     1,
		},
		{
			name:          "user1 starred notebooks ordered by count",
			viewerID:      user1.ID,
			args:          map[string]any{"first": 3, "starredByUserID": graphqlbackend.MarshalUserID(user1.ID), "orderBy": graphqlbackend.NotebookOrderByStarCount, "descending": true},
			wantNotebooks: []*notebooks.Notebook{createdNotebooks[2], createdNotebooks[0]},
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
					idToUsername[createdNotebook.UpdaterUserID],
					createdNotebook.CreatorUserID == tt.viewerID,
				)
				compareNotebookAPIResponses(t, wantNotebookResponse, response.Notebooks.Nodes[idx], true)
			}
		})
	}
}

func TestGetNotebookWithSoftDeletedUserColumns(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	internalCtx := actor.WithInternalActor(context.Background())
	u := db.Users()
	n := notebooks.Notebooks(db)

	user1, err := u.Create(internalCtx, database.NewUser{Username: "u1", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	user2, err := u.Create(internalCtx, database.NewUser{Username: "u2", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	createdNotebook, err := n.CreateNotebook(internalCtx, userNotebookFixture(user2.ID, true))
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	err = u.Delete(internalCtx, user2.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	schema, err := graphqlbackend.NewSchemaWithNotebooksResolver(db, NewResolver(db))
	if err != nil {
		t.Fatal(err)
	}

	input := map[string]any{"id": marshalNotebookID(createdNotebook.ID)}
	var response struct{ Node notebooksapitest.Notebook }
	apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(user1.ID)), t, schema, input, &response, queryNotebook)
}

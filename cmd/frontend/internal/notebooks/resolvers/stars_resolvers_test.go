package resolvers

import (
	"context"
	"fmt"
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
)

const notebookStarFields = `
	user {
		username
	}
	createdAt
`

var createNotebookStarMutation = fmt.Sprintf(`
mutation CreateNotebookStar($notebookID: ID!) {
	createNotebookStar(notebookID: $notebookID) {
		%s
	}
}
`, notebookStarFields)

var deleteNotebookStarMutation = `
mutation DeleteNotebookStar($notebookID: ID!) {
	deleteNotebookStar(notebookID: $notebookID) {
		alwaysNil
	}
}
`

var listNotebookStarsQuery = fmt.Sprintf(`
query NotebookStars($id: ID!, $first: Int!, $after: String) {
	node(id: $id) {
		... on Notebook {
			stars(first: $first, after: $after) {
				nodes {
					%s
			  	}
			  	pageInfo {
					endCursor
					hasNextPage
				}
				totalCount
			}
		}
	}
}
`, notebookStarFields)

func TestCreateAndDeleteNotebookStars(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	internalCtx := actor.WithInternalActor(context.Background())
	u := db.Users()

	user1, err := u.Create(internalCtx, database.NewUser{Username: "u1", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	user2, err := u.Create(internalCtx, database.NewUser{Username: "u2", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	createdNotebooks := createNotebooks(t, db, []*notebooks.Notebook{userNotebookFixture(user1.ID, true), userNotebookFixture(user1.ID, false)})

	schema, err := graphqlbackend.NewSchemaWithNotebooksResolver(db, NewResolver(db))
	if err != nil {
		t.Fatal(err)
	}

	// Create notebook stars for each user
	createAPINotebookStars(t, schema, createdNotebooks[0].ID, user1.ID, user2.ID)

	// Try creating a duplicate notebook star with user1
	input := map[string]any{"notebookID": marshalNotebookID(createdNotebooks[0].ID)}
	var response struct{ CreateNotebookStar notebooksapitest.NotebookStar }
	apiError := apitest.Exec(actor.WithActor(context.Background(), actor.FromUser(user1.ID)), t, schema, input, &response, createNotebookStarMutation)
	if apiError == nil {
		t.Fatalf("expected error when creating a duplicate notebook star, got nil")
	}

	// user2 cannot create a notebook star for user1's private notebook, since user2 does not have access to it
	input = map[string]any{"notebookID": marshalNotebookID(createdNotebooks[1].ID)}
	apiError = apitest.Exec(actor.WithActor(context.Background(), actor.FromUser(user2.ID)), t, schema, input, &response, createNotebookStarMutation)
	if apiError == nil {
		t.Fatalf("expected error when creating a notebook star for inaccessible notebook, got nil")
	}

	// Delete the notebook star for createdNotebooks[0] and user1
	input = map[string]any{"notebookID": marshalNotebookID(createdNotebooks[0].ID)}
	var deleteResponse struct{}
	apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(user1.ID)), t, schema, input, &deleteResponse, deleteNotebookStarMutation)

	// Verify that only one notebook star remains (createdNotebooks[0] and user2)
	input = map[string]any{"id": marshalNotebookID(createdNotebooks[0].ID), "first": 2}
	var listResponse struct {
		Node struct {
			Stars struct {
				Nodes      []notebooksapitest.NotebookStar
				TotalCount int32
				PageInfo   apitest.PageInfo
			}
		}
	}
	apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(user1.ID)), t, schema, input, &listResponse, listNotebookStarsQuery)

	if listResponse.Node.Stars.TotalCount != 1 {
		t.Fatalf("expected 1 notebook star to remain, got %d", listResponse.Node.Stars.TotalCount)
	}
}

func createAPINotebookStars(t *testing.T, schema *graphql.Schema, notebookID int64, userIDs ...int32) []notebooksapitest.NotebookStar {
	t.Helper()
	createdStars := make([]notebooksapitest.NotebookStar, 0, len(userIDs))
	input := map[string]any{"notebookID": marshalNotebookID(notebookID)}
	for _, userID := range userIDs {
		var response struct{ CreateNotebookStar notebooksapitest.NotebookStar }
		apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(userID)), t, schema, input, &response, createNotebookStarMutation)
		createdStars = append(createdStars, response.CreateNotebookStar)
	}
	return createdStars
}

func TestListNotebookStars(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	internalCtx := actor.WithInternalActor(context.Background())
	u := db.Users()

	user1, err := u.Create(internalCtx, database.NewUser{Username: "u1", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	user2, err := u.Create(internalCtx, database.NewUser{Username: "u2", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	user3, err := u.Create(internalCtx, database.NewUser{Username: "u3", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	schema, err := graphqlbackend.NewSchemaWithNotebooksResolver(db, NewResolver(db))
	if err != nil {
		t.Fatal(err)
	}

	createdNotebooks := createNotebooks(t, db, []*notebooks.Notebook{userNotebookFixture(user1.ID, true)})
	createdStars := createAPINotebookStars(t, schema, createdNotebooks[0].ID, user1.ID, user2.ID, user3.ID)

	tests := []struct {
		name      string
		args      map[string]any
		wantCount int32
		wantStars []notebooksapitest.NotebookStar
	}{
		{
			name:      "fetch all notebook stars",
			args:      map[string]any{"id": marshalNotebookID(createdNotebooks[0].ID), "first": 3},
			wantStars: []notebooksapitest.NotebookStar{createdStars[2], createdStars[1], createdStars[0]},
			wantCount: 3,
		},
		{
			name:      "list second page of notebook stars",
			args:      map[string]any{"id": marshalNotebookID(createdNotebooks[0].ID), "first": 1, "after": marshalNotebookStarCursor(1)},
			wantStars: []notebooksapitest.NotebookStar{createdStars[1]},
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var listResponse struct {
				Node struct {
					Stars struct {
						Nodes      []notebooksapitest.NotebookStar
						TotalCount int32
						PageInfo   apitest.PageInfo
					}
				}
			}
			apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(user1.ID)), t, schema, tt.args, &listResponse, listNotebookStarsQuery)

			if tt.wantCount != listResponse.Node.Stars.TotalCount {
				t.Fatalf("expected %d total stars, got %d", tt.wantCount, listResponse.Node.Stars.TotalCount)
			}

			if diff := cmp.Diff(listResponse.Node.Stars.Nodes, tt.wantStars); diff != "" {
				t.Fatalf("wrong notebook stars: %s", diff)
			}
		})
	}
}

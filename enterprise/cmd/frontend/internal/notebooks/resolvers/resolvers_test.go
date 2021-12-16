package resolvers

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/resolvers/apitest"
	notebooksapitest "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/notebooks/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/notebooks"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

const queryNotebook = `
query Notebook($id: ID!) {
	node(id: $id) {
		... on Notebook {
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
		}
	}
}
`

func TestGetNotebook(t *testing.T) {
	t.Parallel()
	db := dbtest.NewDB(t)
	ctx := actor.WithInternalActor(context.Background())
	u := database.Users(db)
	n := notebooks.Notebooks(db)

	user, err := u.Create(ctx, database.NewUser{Username: "u", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

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
	notebook := &notebooks.Notebook{Title: "Notebook Title", Blocks: blocks, Public: true, CreatorUserID: user.ID}
	createdNotebook, err := n.CreateNotebook(ctx, notebook)
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

	wantNotebookResponse := notebooksapitest.Notebook{
		ID:              string(notebookGQLID),
		Title:           createdNotebook.Title,
		Creator:         notebooksapitest.NotebookCreator{Username: "u"},
		CreatedAt:       createdNotebook.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       createdNotebook.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Public:          createdNotebook.Public,
		ViewerCanManage: true,
		Blocks: []notebooksapitest.NotebookBlock{
			{Typename: "QueryBlock", ID: blocks[0].ID, QueryInput: blocks[0].QueryInput.Text},
			{Typename: "MarkdownBlock", ID: blocks[1].ID, MarkdownInput: blocks[1].MarkdownInput.Text},
			{Typename: "FileBlock", ID: blocks[2].ID, FileInput: notebooksapitest.FileInput{
				RepositoryName: blocks[2].FileInput.RepositoryName,
				FilePath:       blocks[2].FileInput.FilePath,
				Revision:       blocks[2].FileInput.Revision,
				LineRange:      &notebooksapitest.LineRange{StartLine: blocks[2].FileInput.LineRange.StartLine, EndLine: blocks[2].FileInput.LineRange.EndLine},
			}},
		},
	}

	if diff := cmp.Diff(wantNotebookResponse, response.Node); diff != "" {
		t.Fatalf("wrong notebook response (-want +got):\n%s", diff)
	}
}

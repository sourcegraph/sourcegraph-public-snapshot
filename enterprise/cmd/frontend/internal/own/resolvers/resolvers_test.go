package resolvers_test

import (
	"context"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/own/resolvers"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/fakedb"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/own/codeowners"
	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/v1"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type fakeOwnService struct{}

func (s fakeOwnService) OwnersFile(context.Context, api.RepoName, api.CommitID) (*codeownerspb.File, error) {
	return &codeownerspb.File{
		Rule: []*codeownerspb.Rule{
			{
				Pattern: "*.js",
				Owner: []*codeownerspb.Owner{
					{Handle: "js-owner"},
				},
			},
		},
	}, nil
}

func (s fakeOwnService) ResolveOwnersWithType(context.Context, []*codeownerspb.Owner) ([]codeowners.ResolvedOwner, error) {
	return nil, nil
}

func userCtx(userID int32) context.Context {
	a := &actor.Actor{
		UID: userID,
	}
	return actor.WithActor(context.Background(), a)
}

type fakeGitserver struct {
	gitserver.Client
}

type fileInfo struct {
	path  string
	size  int64
	isDir bool
}

func (f fileInfo) Name() string { return f.path }
func (f fileInfo) Size() int64  { return f.size }
func (f fileInfo) IsDir() bool  { return f.isDir }
func (f fileInfo) Mode() os.FileMode {
	if f.IsDir() {
		return os.ModeDir
	}
	return 0
}
func (f fileInfo) ModTime() time.Time { return time.Now() }
func (f fileInfo) Sys() any           { return any(nil) }
func (g fakeGitserver) Stat(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, path string) (fs.FileInfo, error) {
	return fileInfo{path: path, size: 42, isDir: false}, nil
}

func TestOwnershipPanelQuery(t *testing.T) {
	fs := fakedb.New()
	db := database.NewMockDB()
	fs.Wire(db)
	var own fakeOwnService
	ctx := userCtx(fs.AddUser(types.User{SiteAdmin: true}))
	repos := database.NewMockRepoStore()
	db.ReposFunc.SetDefaultReturn(repos)
	repos.GetFunc.PushReturn(&types.Repo{}, nil)
	backend.Mocks.Repos.ResolveRev = func(_ context.Context, repo *types.Repo, rev string) (api.CommitID, error) {
		return "42", nil
	}
	git := fakeGitserver{}
	schema, err := graphqlbackend.NewSchema(db, git, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, resolvers.New(db, own))
	if err != nil {
		t.Fatal(err)
	}
	graphqlbackend.RunTest(t, &graphqlbackend.Test{
		Schema:  schema,
		Context: ctx,
		Query: `
			fragment OwnerFields on Person {
				email
				avatarURL
				displayName
				user {
					username
					displayName
					url
				}
			}

			fragment CodeownersFileEntryFields on CodeownersFileEntry {
				title
				description
			}

			query FetchOwnership($repo: ID!, $revision: String!, $currentPath: String!) {
				node(id: $repo) {
					... on Repository {
						commit(rev: $revision) {
							blob(path: $currentPath) {
								ownership {
									nodes {
										owner {
											...OwnerFields
										}
										reasons {
											...CodeownersFileEntryFields
										}
									}
								}
							}
						}
					}
				}
			}`,
		ExpectedResult: `{
			"node": {
				"commit": {
					"blob": {
						"ownership": {
							"nodes": [
								{
									"owner": {
									},
									"reasons": {
									}
								}
							]
						}
					}
				}
			}
		}`,
		Variables: map[string]any{
			"repo":        string(relay.MarshalID("Repository", 42)),
			"revision":    "revision",
			"currentPath": "foo/bar.js",
		},
	})
}

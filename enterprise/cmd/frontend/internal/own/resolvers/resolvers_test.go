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
	"github.com/sourcegraph/sourcegraph/internal/types"

	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/v1"
)

// userCtx returns a context where give user ID identifies logged in user.
func userCtx(userID int32) context.Context {
	a := &actor.Actor{
		UID: userID,
	}
	return actor.WithActor(context.Background(), a)
}

// fakeOwnService returns given owners file and resolves owners to UnknownOwner.
type fakeOwnService struct {
	File *codeownerspb.File
}

func (s fakeOwnService) OwnersFile(context.Context, api.RepoName, api.CommitID) (*codeownerspb.File, error) {
	return s.File, nil
}

// ResolverOwnersWithType here behaves in line with production
// OwnService implementation in case handle/email cannot be associated
// with anything - defaults to a Person with a nil person entity.
func (s fakeOwnService) ResolveOwnersWithType(_ context.Context, owners []*codeownerspb.Owner) ([]codeowners.ResolvedOwner, error) {
	var resolved []codeowners.ResolvedOwner
	for _, o := range owners {
		resolved = append(resolved, &codeowners.Person{
			Handle: o.Handle,
			Email:  o.Email,
		})
	}
	return resolved, nil
}

// fakeGitServer is a limited gitserver.Client that returns a file for every Stat call.
type fakeGitserver struct {
	gitserver.Client
}

// Stat is a fake implementation that returns a FileInfo
// indicating a regular file for every path it is given.
func (g fakeGitserver) Stat(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, path string) (fs.FileInfo, error) {
	return fileInfo{path: path}, nil
}

// fileInfo fakes fs.FileInfo for given path, and pretends to be a regular file.
type fileInfo struct {
	path string
}

func (f fileInfo) Name() string { return f.path }
func (f fileInfo) Size() int64  { return 0 }
func (f fileInfo) IsDir() bool  { return false }
func (f fileInfo) Mode() os.FileMode {
	if f.IsDir() {
		return os.ModeDir
	}
	return 0
}
func (f fileInfo) ModTime() time.Time { return time.Now() }
func (f fileInfo) Sys() any           { return any(nil) }

// TestBlobOwnershipPanelQueryPersonUnresolved mimics the blob ownership panel graphQL
// query, where the owner is unresolved. In that case if we have a handle, we only return
// it as `displayName`.
func TestBlobOwnershipPanelQueryPersonUnresolved(t *testing.T) {
	fs := fakedb.New()
	db := database.NewMockDB()
	fs.Wire(db)
	own := fakeOwnService{
		File: &codeownerspb.File{
			Rule: []*codeownerspb.Rule{
				{
					Pattern: "*.js",
					Owner: []*codeownerspb.Owner{
						{Handle: "js-owner"},
					},
				},
			},
		},
	}
	ctx := userCtx(fs.AddUser(types.User{SiteAdmin: true}))
	repos := database.NewMockRepoStore()
	db.ReposFunc.SetDefaultReturn(repos)
	repos.GetFunc.SetDefaultReturn(&types.Repo{}, nil)
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
										"email": "",
										"avatarURL": null,
										"displayName": "js-owner",
										"user": null
									},
									"reasons": [
										{
											"title": "CodeOwners",
											"description": "Dummy code owners file entry."
										}
									]
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

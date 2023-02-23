package resolvers_test

import (
	"context"
	"io/fs"
	"testing"

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
	ctx := context.Background()
	a := actor.FromUser(userID)
	return actor.WithActor(ctx, a)
}

// fakeOwnershipGraph returns ownership according to given owners file and resolves
// each owners to Person with just email or handle as in the File rules.
type fakeOwnershipGraph struct {
	backend.OwnService
	File *codeownerspb.File
}

func (s fakeOwnershipGraph) Ownership(context.Context, api.RepoName, api.CommitID) (codeowners.Graph, error) {
	return s, nil
}

func (s fakeOwnershipGraph) FindOwners(_ context.Context, filePath string) ([]codeowners.Ownership, error) {
	var ownerships []codeowners.Ownership
	for _, r := range s.File.GetRule() {
		glob, err := codeownerspb.Compile(r.GetPattern())
		if err != nil {
			return nil, err
		}
		if glob.Match(filePath) {
			for _, o := range r.GetOwner() {
				ownerships = append(ownerships, codeowners.Ownership{
					Owner: &codeowners.Person{
						Handle: o.Handle,
						Email:  o.Email,
					},
					Rule: r,
				})
			}
		}
	}
	return ownerships, nil
}

// fakeGitServer is a limited gitserver.Client that returns a file for every Stat call.
type fakeGitserver struct {
	gitserver.Client
}

// Stat is a fake implementation that returns a FileInfo
// indicating a regular file for every path it is given.
func (g fakeGitserver) Stat(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, path string) (fs.FileInfo, error) {
	return graphqlbackend.CreateFileInfo(path, false), nil
}

// TestBlobOwnershipPanelQueryPersonUnresolved mimics the blob ownership panel graphQL
// query, where the owner is unresolved. In that case if we have a handle, we only return
// it as `displayName`.
func TestBlobOwnershipPanelQueryPersonUnresolved(t *testing.T) {
	fs := fakedb.New()
	db := database.NewMockDB()
	fs.Wire(db)
	own := fakeOwnershipGraph{
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
											"description": "Owner is associated with a rule in code owners file."
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

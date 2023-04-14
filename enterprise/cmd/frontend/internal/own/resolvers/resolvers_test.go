package resolvers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/own/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/own"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/own/codeowners"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/fakedb"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	enterprisedb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	codeownerspb "github.com/sourcegraph/sourcegraph/enterprise/internal/own/codeowners/v1"
)

// userCtx returns a context where give user ID identifies logged in user.
func userCtx(userID int32) context.Context {
	ctx := context.Background()
	a := actor.FromUser(userID)
	return actor.WithActor(ctx, a)
}

// fakeOwnService returns given owners file and resolves owners to UnknownOwner.
type fakeOwnService struct {
	Ruleset *codeowners.Ruleset
}

func (s fakeOwnService) RulesetForRepo(context.Context, api.RepoName, api.RepoID, api.CommitID) (*codeowners.Ruleset, error) {
	return s.Ruleset, nil
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
	files repoFiles
}

type repoPath struct {
	Repo     api.RepoName
	CommitID api.CommitID
	Path     string
}

type repoFiles map[repoPath]string

func (g fakeGitserver) ReadFile(_ context.Context, _ authz.SubRepoPermissionChecker, repoName api.RepoName, commitID api.CommitID, file string) ([]byte, error) {
	if g.files == nil {
		return nil, os.ErrNotExist
	}
	content, ok := g.files[repoPath{Repo: repoName, CommitID: commitID, Path: file}]
	if !ok {
		return nil, os.ErrNotExist
	}
	return []byte(content), nil
}

// Stat is a fake implementation that returns a FileInfo
// indicating a regular file for every path it is given.
func (g fakeGitserver) Stat(_ context.Context, _ authz.SubRepoPermissionChecker, _ api.RepoName, _ api.CommitID, path string) (fs.FileInfo, error) {
	return graphqlbackend.CreateFileInfo(path, false), nil
}

// TestBlobOwnershipPanelQueryPersonUnresolved mimics the blob ownership panel graphQL
// query, where the owner is unresolved. In that case if we have a handle, we only return
// it as `displayName`.
func TestBlobOwnershipPanelQueryPersonUnresolved(t *testing.T) {
	logger := logtest.Scoped(t)
	fakeDB := fakedb.New()
	db := database.NewMockDB()
	fakeDB.Wire(db)
	repoID := api.RepoID(1)
	own := fakeOwnService{
		Ruleset: codeowners.NewRuleset(
			codeowners.GitRulesetSource{Repo: repoID, Commit: "deadbeef", Path: "CODEOWNERS"},
			&codeownerspb.File{
				Rule: []*codeownerspb.Rule{
					{
						Pattern: "*.js",
						Owner: []*codeownerspb.Owner{
							{Handle: "js-owner"},
						},
						LineNumber: 1,
					},
				},
			}),
	}
	ctx := userCtx(fakeDB.AddUser(types.User{SiteAdmin: true}))
	ctx = featureflag.WithFlags(ctx, featureflag.NewMemoryStore(map[string]bool{"search-ownership": true}, nil, nil))
	repos := database.NewMockRepoStore()
	db.ReposFunc.SetDefaultReturn(repos)
	repos.GetFunc.SetDefaultReturn(&types.Repo{ID: repoID, Name: "github.com/sourcegraph/own"}, nil)
	backend.Mocks.Repos.ResolveRev = func(_ context.Context, repo *types.Repo, rev string) (api.CommitID, error) {
		return "deadbeef", nil
	}
	git := fakeGitserver{}
	schema, err := graphqlbackend.NewSchema(db, git, nil, graphqlbackend.OptionalResolver{OwnResolver: resolvers.NewWithService(db, git, own, logger)})
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
				codeownersFile {
					__typename
					url
				}
				ruleLineMatch
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
											"title": "CODEOWNERS",
											"description": "Owner is associated with a rule in a CODEOWNERS file.",
											"codeownersFile": {
												"__typename": "GitBlob",
												"url": "/github.com/sourcegraph/own@deadbeef/-/blob/CODEOWNERS"
											},
											"ruleLineMatch": 1
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

func TestBlobOwnershipPanelQueryIngested(t *testing.T) {
	logger := logtest.Scoped(t)
	fakeDB := fakedb.New()
	db := database.NewMockDB()
	fakeDB.Wire(db)
	repoID := api.RepoID(1)
	own := fakeOwnService{
		Ruleset: codeowners.NewRuleset(
			codeowners.IngestedRulesetSource{ID: int32(repoID)},
			&codeownerspb.File{
				Rule: []*codeownerspb.Rule{
					{
						Pattern: "*.js",
						Owner: []*codeownerspb.Owner{
							{Handle: "js-owner"},
						},
						LineNumber: 1,
					},
				},
			}),
	}
	ctx := userCtx(fakeDB.AddUser(types.User{SiteAdmin: true}))
	ctx = featureflag.WithFlags(ctx, featureflag.NewMemoryStore(map[string]bool{"search-ownership": true}, nil, nil))
	repos := database.NewMockRepoStore()
	db.ReposFunc.SetDefaultReturn(repos)
	repos.GetFunc.SetDefaultReturn(&types.Repo{ID: repoID, Name: "github.com/sourcegraph/own"}, nil)
	backend.Mocks.Repos.ResolveRev = func(_ context.Context, repo *types.Repo, rev string) (api.CommitID, error) {
		return "deadbeef", nil
	}
	git := fakeGitserver{}
	schema, err := graphqlbackend.NewSchema(db, git, nil, graphqlbackend.OptionalResolver{OwnResolver: resolvers.NewWithService(db, git, own, logger)})
	if err != nil {
		t.Fatal(err)
	}
	graphqlbackend.RunTest(t, &graphqlbackend.Test{
		Schema:  schema,
		Context: ctx,
		Query: `
			fragment CodeownersFileEntryFields on CodeownersFileEntry {
				title
				description
				codeownersFile {
					__typename
					url
				}
				ruleLineMatch
			}

			query FetchOwnership($repo: ID!, $revision: String!, $currentPath: String!) {
				node(id: $repo) {
					... on Repository {
						commit(rev: $revision) {
							blob(path: $currentPath) {
								ownership {
									nodes {
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
									"reasons": [
										{
											"title": "CODEOWNERS",
											"description": "Owner is associated with a rule in a CODEOWNERS file.",
											"codeownersFile": {
												"__typename": "VirtualFile",
												"url": "/github.com/sourcegraph/own/-/own"
											},
											"ruleLineMatch": 1
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
			"repo":        string(relay.MarshalID("Repository", repoID)),
			"revision":    "revision",
			"currentPath": "foo/bar.js",
		},
	})
}

func TestBlobOwnershipPanelQueryTeamResolved(t *testing.T) {
	logger := logtest.Scoped(t)
	repo := &types.Repo{Name: "repo-name", ID: 42}
	team := &types.Team{Name: "fake-team", DisplayName: "The Fake Team"}
	var parameterRevision = "revision-parameter"
	var resolvedRevision api.CommitID = "revision-resolved"
	git := fakeGitserver{
		files: repoFiles{
			{repo.Name, resolvedRevision, "CODEOWNERS"}: "*.js @fake-team",
		},
	}
	fakeDB := fakedb.New()
	db := enterprisedb.NewMockEnterpriseDB()
	db.TeamsFunc.SetDefaultReturn(fakeDB.TeamStore)
	db.UsersFunc.SetDefaultReturn(fakeDB.UserStore)
	db.CodeownersFunc.SetDefaultReturn(enterprisedb.NewMockCodeownersStore())
	own := own.NewService(git, db)
	ctx := userCtx(fakeDB.AddUser(types.User{SiteAdmin: true}))
	ctx = featureflag.WithFlags(ctx, featureflag.NewMemoryStore(map[string]bool{"search-ownership": true}, nil, nil))
	repos := database.NewMockRepoStore()
	db.ReposFunc.SetDefaultReturn(repos)
	repos.GetFunc.SetDefaultReturn(repo, nil)
	backend.Mocks.Repos.ResolveRev = func(_ context.Context, repo *types.Repo, rev string) (api.CommitID, error) {
		if rev != parameterRevision {
			return "", errors.Newf("ResolveRev, got %q want %q", rev, parameterRevision)
		}
		return resolvedRevision, nil
	}
	if err := fakeDB.TeamStore.CreateTeam(ctx, team); err != nil {
		t.Fatalf("failed to create fake team: %s", err)
	}
	schema, err := graphqlbackend.NewSchema(db, git, nil, graphqlbackend.OptionalResolver{OwnResolver: resolvers.NewWithService(db, git, own, logger)})
	if err != nil {
		t.Fatal(err)
	}
	graphqlbackend.RunTest(t, &graphqlbackend.Test{
		Schema:  schema,
		Context: ctx,
		Query: `
			query FetchOwnership($repo: ID!, $revision: String!, $currentPath: String!) {
				node(id: $repo) {
					... on Repository {
						commit(rev: $revision) {
							blob(path: $currentPath) {
								ownership {
									nodes {
										owner {
											... on Team {
												displayName
											}
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
										"displayName": "The Fake Team"
									}
								}
							]
						}
					}
				}
			}
		}`,
		Variables: map[string]any{
			"repo":        string(relay.MarshalID("Repository", int(repo.ID))),
			"revision":    parameterRevision,
			"currentPath": "foo/bar.js",
		},
	})
}

var paginationQuery = `
query FetchOwnership($repo: ID!, $revision: String!, $currentPath: String!, $after: String!) {
	node(id: $repo) {
		... on Repository {
			commit(rev: $revision) {
				blob(path: $currentPath) {
					ownership(first: 2, after: $after) {
						totalCount
						pageInfo {
							endCursor
							hasNextPage
						}
						nodes {
							owner {
								...on Person {
									displayName
								}
							}
						}
					}
				}
			}
		}
	}
}`

type paginationResponse struct {
	Node struct {
		Commit struct {
			Blob struct {
				Ownership struct {
					TotalCount int
					PageInfo   struct {
						EndCursor   *string
						HasNextPage bool
					}
					Nodes []struct {
						Owner struct {
							DisplayName string
						}
					}
				}
			}
		}
	}
}

func (r paginationResponse) hasNextPage() bool {
	return r.Node.Commit.Blob.Ownership.PageInfo.HasNextPage
}

func (r paginationResponse) consistentPageInfo() error {
	ownership := r.Node.Commit.Blob.Ownership
	if nextPage, hasCursor := ownership.PageInfo.HasNextPage, ownership.PageInfo.EndCursor != nil; nextPage != hasCursor {
		cursor := "<nil>"
		if ownership.PageInfo.EndCursor != nil {
			cursor = fmt.Sprintf("&%q", *ownership.PageInfo.EndCursor)
		}
		return errors.Newf("PageInfo.HasNextPage %v but PageInfo.EndCursor %s", nextPage, cursor)
	}
	return nil
}

func (r paginationResponse) ownerNames() []string {
	var owners []string
	for _, n := range r.Node.Commit.Blob.Ownership.Nodes {
		owners = append(owners, n.Owner.DisplayName)
	}
	return owners
}

// TestOwnershipPagination issues a number of queries using ownership(first) parameter
// to limit number of responses. It expects to see correct pagination behavior, that is:
// *  all results are eventually returned, in the expected order;
// *  each request returns correct pageInfo and totalCount;
func TestOwnershipPagination(t *testing.T) {
	logger := logtest.Scoped(t)
	fakeDB := fakedb.New()
	db := database.NewMockDB()
	fakeDB.Wire(db)
	rule := &codeownerspb.Rule{
		Pattern: "*.js",
		Owner: []*codeownerspb.Owner{
			{Handle: "js-owner-1"},
			{Handle: "js-owner-2"},
			{Handle: "js-owner-3"},
			{Handle: "js-owner-4"},
			{Handle: "js-owner-5"},
		},
	}

	own := fakeOwnService{
		Ruleset: codeowners.NewRuleset(
			codeowners.IngestedRulesetSource{},
			&codeownerspb.File{
				Rule: []*codeownerspb.Rule{rule},
			}),
	}
	ctx := userCtx(fakeDB.AddUser(types.User{SiteAdmin: true}))
	ctx = featureflag.WithFlags(ctx, featureflag.NewMemoryStore(map[string]bool{"search-ownership": true}, nil, nil))
	repos := database.NewMockRepoStore()
	db.ReposFunc.SetDefaultReturn(repos)
	repos.GetFunc.SetDefaultReturn(&types.Repo{}, nil)
	backend.Mocks.Repos.ResolveRev = func(_ context.Context, repo *types.Repo, rev string) (api.CommitID, error) {
		return "42", nil
	}
	git := fakeGitserver{}
	schema, err := graphqlbackend.NewSchema(db, git, nil, graphqlbackend.OptionalResolver{OwnResolver: resolvers.NewWithService(db, git, own, logger)})
	if err != nil {
		t.Fatal(err)
	}
	var after string
	var paginatedOwners [][]string
	var lastResponseData *paginationResponse
	// Limit iterations to number of owners total, so that the test
	// has a stop condition in case something malfunctions.
	for i := 0; i < len(rule.Owner); i++ {
		var responseData paginationResponse
		variables := map[string]any{
			"repo":        string(relay.MarshalID("Repository", 42)),
			"revision":    "revision",
			"currentPath": "foo/bar.js",
			"after":       after,
		}
		response := schema.Exec(ctx, paginationQuery, "", variables)
		for _, err := range response.Errors {
			t.Errorf("GraphQL Exec, errors: %s", err)
		}
		if response.Data == nil {
			t.Fatal("GraphQL response has no data.")
		}
		if err := json.Unmarshal(response.Data, &responseData); err != nil {
			t.Fatalf("Cannot unmarshal GrapgQL JSON response: %s", err)
		}
		ownership := responseData.Node.Commit.Blob.Ownership
		if got, want := ownership.TotalCount, len(rule.Owner); got != want {
			t.Errorf("TotalCount, got %d want %d", got, want)
		}
		paginatedOwners = append(paginatedOwners, responseData.ownerNames())
		if err := responseData.consistentPageInfo(); err != nil {
			t.Error(err)
		}
		lastResponseData = &responseData
		if ownership.PageInfo.HasNextPage {
			after = *ownership.PageInfo.EndCursor
		} else {
			break
		}
	}
	if lastResponseData == nil {
		t.Error("No response received.")
	} else if lastResponseData.hasNextPage() {
		t.Error("Last response has next page information - result is not exhaustive.")
	}
	wantPaginatedOwners := [][]string{
		{
			"js-owner-1",
			"js-owner-2",
		},
		{
			"js-owner-3",
			"js-owner-4",
		},
		{
			"js-owner-5",
		},
	}
	if diff := cmp.Diff(wantPaginatedOwners, paginatedOwners); diff != "" {
		t.Errorf("returned owners -want+got: %s", diff)
	}
}

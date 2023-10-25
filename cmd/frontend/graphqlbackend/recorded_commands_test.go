package graphqlbackend

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	gqlerrors "github.com/graph-gophers/graphql-go/errors"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/database/fakedb"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestRecordedCommandsResolver(t *testing.T) {
	rcache.SetupForTest(t)

	timeFormat := "2006-01-02T15:04:05Z"
	startTime, err := time.Parse(timeFormat, "2023-07-20T15:04:05Z")
	require.NoError(t, err)

	fs := fakedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{Username: "bob", SiteAdmin: true})
	ctx := userCtx(userID)

	repoName := "github.com/sourcegraph/sourcegraph"
	backend.Mocks.Repos.GetByName = func(context.Context, api.RepoName) (*types.Repo, error) {
		return &types.Repo{Name: api.RepoName(repoName)}, nil
	}
	t.Cleanup(func() {
		backend.Mocks = backend.MockServices{}
	})

	t.Run("gitRecoreder not configured for repository", func(t *testing.T) {
		// When gitRecorder isn't set, we return an empty list.
		RunTest(t, &Test{
			Schema:  mustParseGraphQLSchema(t, db),
			Context: ctx,
			Query: `
				{
					repository(name: "github.com/sourcegraph/sourcegraph") {
						recordedCommands {
							nodes {
								start
								duration
								command
								dir
								path
							}
							totalCount
							pageInfo {
								hasNextPage
							}
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"repository": {
						"recordedCommands": {
							"nodes": [],
							"totalCount": 0,
							"pageInfo": {
								"hasNextPage": false
							}
						}
					}
				}
			`,
		})

	})

	t.Run("no recorded commands for repository", func(t *testing.T) {
		conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{GitRecorder: &schema.GitRecorder{Size: 3}}})
		t.Cleanup(func() { conf.Mock(nil) })

		repos := dbmocks.NewMockRepoStore()
		repos.GetFunc.SetDefaultReturn(&types.Repo{Name: api.RepoName(repoName)}, nil)
		db.ReposFunc.SetDefaultReturn(repos)

		RunTest(t, &Test{
			Schema:  mustParseGraphQLSchema(t, db),
			Context: ctx,
			Query: `
					{
						repository(name: "github.com/sourcegraph/sourcegraph") {
							recordedCommands {
								nodes {
									start
									duration
									command
									dir
									path
								}
								totalCount
								pageInfo {
									hasNextPage
								}
							}
						}
					}
				`,
			ExpectedResult: `
					{
						"repository": {
							"recordedCommands": {
								"nodes": [],
								"totalCount": 0,
								"pageInfo": {
									"hasNextPage": false
								}
							}
						}
					}
				`,
		})

	})

	t.Run("one recorded command for repository", func(t *testing.T) {
		conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{GitRecorder: &schema.GitRecorder{Size: 3}}})
		t.Cleanup(func() { conf.Mock(nil) })

		repos := dbmocks.NewMockRepoStore()
		repos.GetFunc.SetDefaultReturn(&types.Repo{Name: api.RepoName(repoName)}, nil)
		db.ReposFunc.SetDefaultReturn(repos)

		r := rcache.NewFIFOList(wrexec.GetFIFOListKey(repoName), 3)
		cmd1 := wrexec.RecordedCommand{
			Start:    startTime,
			Duration: float64(100),
			Args:     []string{"git", "fetch"},
			Dir:      "/.sourcegraph/repos_1/github.com/sourcegraph/sourcegraph/.git",
			Path:     "/opt/homebrew/bin/git",
		}
		err = r.Insert(marshalCmd(t, cmd1))
		require.NoError(t, err)

		RunTest(t, &Test{
			Schema:  mustParseGraphQLSchema(t, db),
			Context: ctx,
			Query: `
				{
					repository(name: "github.com/sourcegraph/sourcegraph") {
						recordedCommands {
							nodes {
								start
								duration
								command
								dir
								path
							}
							totalCount
							pageInfo {
								hasNextPage
							}
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"repository": {
						"recordedCommands": {
							"nodes": [
								{
									"command": "git fetch",
									"dir": "/.sourcegraph/repos_1/github.com/sourcegraph/sourcegraph/.git",
									"duration": 100,
									"path": "/opt/homebrew/bin/git",
									"start": "2023-07-20T15:04:05Z"
								}
							],
							"totalCount": 1,
							"pageInfo": {
								"hasNextPage": false
							}
						}
					}
				}
			`,
		})

	})

	t.Run("paginated recorded commands", func(t *testing.T) {
		cmd1 := wrexec.RecordedCommand{
			Start:    startTime,
			Duration: float64(100),
			Args:     []string{"git", "fetch"},
			Dir:      "/.sourcegraph/repos_1/github.com/sourcegraph/sourcegraph/.git",
			Path:     "/opt/homebrew/bin/git",
		}
		cmd2 := wrexec.RecordedCommand{
			Start:    startTime,
			Duration: float64(10),
			Args:     []string{"git", "clone"},
			Dir:      "/.sourcegraph/repos_1/github.com/sourcegraph/sourcegraph/.git",
			Path:     "/opt/homebrew/bin/git",
		}
		cmd3 := wrexec.RecordedCommand{
			Start:    startTime,
			Duration: float64(5),
			Args:     []string{"git", "ls-files"},
			Dir:      "/.sourcegraph/repos_1/github.com/sourcegraph/sourcegraph/.git",
			Path:     "/opt/homebrew/bin/git",
		}

		conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{GitRecorder: &schema.GitRecorder{Size: 3}}})
		t.Cleanup(func() { conf.Mock(nil) })

		repos := dbmocks.NewMockRepoStore()
		repos.GetFunc.SetDefaultReturn(&types.Repo{Name: api.RepoName(repoName)}, nil)
		db.ReposFunc.SetDefaultReturn(repos)

		r := rcache.NewFIFOList(wrexec.GetFIFOListKey(repoName), 3)

		err = r.Insert(marshalCmd(t, cmd1))
		require.NoError(t, err)
		err = r.Insert(marshalCmd(t, cmd2))
		require.NoError(t, err)
		err = r.Insert(marshalCmd(t, cmd3))
		require.NoError(t, err)

		t.Run("limit within bounds", func(t *testing.T) {
			RunTest(t, &Test{
				Schema:  mustParseGraphQLSchema(t, db),
				Context: ctx,
				Query: `
						{
							repository(name: "github.com/sourcegraph/sourcegraph") {
								recordedCommands(limit: 2) {
									nodes {
										start
										duration
										command
										dir
										path
									}
									totalCount
									pageInfo {
										hasNextPage
									}
								}
							}
						}
					`,
				ExpectedResult: `
						{
							"repository": {
								"recordedCommands": {
									"nodes": [
										{
											"command": "git ls-files",
											"dir": "/.sourcegraph/repos_1/github.com/sourcegraph/sourcegraph/.git",
											"duration": 5,
											"path": "/opt/homebrew/bin/git",
											"start": "2023-07-20T15:04:05Z"
										},
										{
											"command": "git clone",
											"dir": "/.sourcegraph/repos_1/github.com/sourcegraph/sourcegraph/.git",
											"duration": 10,
											"path": "/opt/homebrew/bin/git",
											"start": "2023-07-20T15:04:05Z"
										}
									],
									"totalCount": 3,
									"pageInfo": {
										"hasNextPage": true
									}
								}
							}
						}
					`,
			})
		})

		t.Run("limit exceeds bounds", func(t *testing.T) {
			RunTest(t, &Test{
				Schema:  mustParseGraphQLSchema(t, db),
				Context: ctx,
				Query: `
						{
							repository(name: "github.com/sourcegraph/sourcegraph") {
								recordedCommands(limit: 10000) {
									nodes {
										start
										duration
										command
										dir
										path
									}
									totalCount
									pageInfo {
										hasNextPage
									}
								}
							}
						}
					`,
				ExpectedResult: `
						{
							"repository": {
								"recordedCommands": {
									"nodes": [
										{
											"command": "git ls-files",
											"dir": "/.sourcegraph/repos_1/github.com/sourcegraph/sourcegraph/.git",
											"duration": 5,
											"path": "/opt/homebrew/bin/git",
											"start": "2023-07-20T15:04:05Z"
										},
										{
											"command": "git clone",
											"dir": "/.sourcegraph/repos_1/github.com/sourcegraph/sourcegraph/.git",
											"duration": 10,
											"path": "/opt/homebrew/bin/git",
											"start": "2023-07-20T15:04:05Z"
										},
										{
											"command": "git fetch",
											"dir": "/.sourcegraph/repos_1/github.com/sourcegraph/sourcegraph/.git",
											"duration": 100,
											"path": "/opt/homebrew/bin/git",
											"start": "2023-07-20T15:04:05Z"
										}
									],
									"totalCount": 3,
									"pageInfo": {
										"hasNextPage": false
									}
								}
							}
						}
					`,
			})
		})

		t.Run("offset exceeds total count", func(t *testing.T) {
			RunTest(t, &Test{
				Schema:  mustParseGraphQLSchema(t, db),
				Context: ctx,
				Query: `
						{
							repository(name: "github.com/sourcegraph/sourcegraph") {
								recordedCommands(offset: 1000) {
									nodes {
										start
										duration
										command
										dir
										path
									}
									totalCount
									pageInfo {
										hasNextPage
									}
								}
							}
						}
					`,
				ExpectedResult: `
						{
							"repository": {
								"recordedCommands": {
									"nodes": [],
									"totalCount": 3,
									"pageInfo": {
										"hasNextPage": false
									}
								}
							}
						}
					`,
			})
		})

		t.Run("valid offset and limit", func(t *testing.T) {
			RunTest(t, &Test{
				Schema:  mustParseGraphQLSchema(t, db),
				Context: ctx,
				Query: `
						{
							repository(name: "github.com/sourcegraph/sourcegraph") {
								recordedCommands(offset: 1, limit: 2) {
									nodes {
										start
										duration
										command
										dir
										path
									}
									totalCount
									pageInfo {
										hasNextPage
									}
								}
							}
						}
					`,
				ExpectedResult: `
						{
							"repository": {
								"recordedCommands": {
									"nodes": [
										{
											"command": "git clone",
											"dir": "/.sourcegraph/repos_1/github.com/sourcegraph/sourcegraph/.git",
											"duration": 10,
											"path": "/opt/homebrew/bin/git",
											"start": "2023-07-20T15:04:05Z"
										},
										{
											"command": "git fetch",
											"dir": "/.sourcegraph/repos_1/github.com/sourcegraph/sourcegraph/.git",
											"duration": 100,
											"path": "/opt/homebrew/bin/git",
											"start": "2023-07-20T15:04:05Z"
										}
									],
									"totalCount": 3,
									"pageInfo": {
										"hasNextPage": false
									}
								}
							}
						}
					`,
			})
		})

		t.Run("limit exceeds recordedCommandMaxLimit", func(t *testing.T) {
			MockGetRecordedCommandMaxLimit = func() int {
				return 1
			}
			t.Cleanup(func() {
				MockGetRecordedCommandMaxLimit = nil
			})
			RunTest(t, &Test{
				Schema:  mustParseGraphQLSchema(t, db),
				Context: ctx,
				Query: `
						{
							repository(name: "github.com/sourcegraph/sourcegraph") {
								recordedCommands(limit: 20) {
									nodes {
										start
										duration
										command
										dir
										path
									}
									totalCount
									pageInfo {
										hasNextPage
									}
								}
							}
						}
					`,
				ExpectedResult: `
						{
							"repository": {
								"recordedCommands": {
									"nodes": [
										{
											"command": "git ls-files",
											"dir": "/.sourcegraph/repos_1/github.com/sourcegraph/sourcegraph/.git",
											"duration": 5,
											"path": "/opt/homebrew/bin/git",
											"start": "2023-07-20T15:04:05Z"
										}
									],
									"totalCount": 3,
									"pageInfo": {
										"hasNextPage": true
									}
								}
							}
						}
					`,
			})
		})
	})

	t.Run("user is not a site-admin", func(t *testing.T) {
		conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{GitRecorder: &schema.GitRecorder{Size: 3}}})
		t.Cleanup(func() { conf.Mock(nil) })

		repos := dbmocks.NewMockRepoStore()
		repos.GetFunc.SetDefaultReturn(&types.Repo{Name: api.RepoName(repoName)}, nil)
		db.ReposFunc.SetDefaultReturn(repos)

		userID := fs.AddUser(types.User{Username: "will", SiteAdmin: false})
		ctx := userCtx(userID)

		RunTest(t, &Test{
			Schema:  mustParseGraphQLSchema(t, db),
			Context: ctx,
			Query: `
					{
						repository(name: "github.com/sourcegraph/sourcegraph") {
							recordedCommands {
								nodes {
									command
								}
							}
						}
					}
				`,
			ExpectedResult: `{"repository": null}`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Message: "must be site admin",
					Path:    []any{string("repository"), string("recordedCommands")},
				},
			},
		})

	})
}

func marshalCmd(t *testing.T, command wrexec.RecordedCommand) []byte {
	t.Helper()
	bytes, err := json.Marshal(&command)
	require.NoError(t, err)
	return bytes
}

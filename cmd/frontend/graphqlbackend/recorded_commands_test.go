package graphqlbackend

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/schema"
	"github.com/stretchr/testify/require"
)

func TestRecordedCommandsResolver(t *testing.T) {
	rcache.SetupForTest(t)

	timeFormat := "2006-01-02T15:04:05Z"
	startTime, err := time.Parse(timeFormat, "2023-07-20T15:04:05Z")
	require.NoError(t, err)

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
	db := database.NewMockDB()

	repoName := "github.com/sourcegraph/sourcegraph"
	backend.Mocks.Repos.GetByName = func(context.Context, api.RepoName) (*types.Repo, error) {
		return &types.Repo{Name: api.RepoName(repoName)}, nil
	}
	t.Cleanup(func() {
		backend.Mocks = backend.MockServices{}
	})

	repos := database.NewMockRepoStore()
	repos.GetFunc.SetDefaultReturn(&types.Repo{Name: api.RepoName(repoName)}, nil)
	db.ReposFunc.SetDefaultReturn(repos)

	RunTest(t, &Test{
		Schema: mustParseGraphQLSchema(t, db),
		Query: `
				{
					repository(name: "github.com/sourcegraph/sourcegraph") {
						recordedCommands {
							start
							duration
							command
							dir
							path
						}
					}
				}
			`,
		ExpectedResult: `
				{
					"repository": {
						"recordedCommands": []
					}
				}
			`,
	})

	r := rcache.NewFIFOList(wrexec.GetFIFOListKey(repoName), 3)
	err = r.Insert(marshalCmd(t, cmd1))
	require.NoError(t, err)

	RunTest(t, &Test{
		Schema: mustParseGraphQLSchema(t, db),
		Query: `
				{
					repository(name: "github.com/sourcegraph/sourcegraph") {
						recordedCommands {
							start
							duration
							command
							dir
							path
						}
					}
				}
			`,
		ExpectedResult: `
				{
					"repository": {
						"recordedCommands": [
							{
								"command": "git fetch",
								"dir": "/.sourcegraph/repos_1/github.com/sourcegraph/sourcegraph/.git",
								"duration": 100,
								"path": "/opt/homebrew/bin/git",
								"start": "2023-07-20T15:04:05Z"
							}
						]
					}
				}
			`,
	})

	err = r.Insert(marshalCmd(t, cmd2))
	require.NoError(t, err)
	err = r.Insert(marshalCmd(t, cmd3))
	require.NoError(t, err)

	RunTest(t, &Test{
		Schema: mustParseGraphQLSchema(t, db),
		Query: `
				{
					repository(name: "github.com/sourcegraph/sourcegraph") {
						recordedCommands {
							start
							duration
							command
							dir
							path
						}
					}
				}
			`,
		ExpectedResult: `
				{
					"repository": {
						"recordedCommands": [
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
						]
					}
				}
			`,
	})
}

func marshalCmd(t *testing.T, command wrexec.RecordedCommand) []byte {
	t.Helper()
	bytes, err := json.Marshal(&command)
	require.NoError(t, err)
	return bytes
}

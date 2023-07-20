package graphqlbackend

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/graph-gophers/graphql-go/errors"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
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

	db := database.NewMockDB()
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{GitRecorder: &schema.GitRecorder{Size: 3}}})
	t.Cleanup(func() { conf.Mock(nil) })

	RunTest(t, &Test{
		Schema: mustParseGraphQLSchema(t, db),
		Query: `
				{
					recordedCommands(repoName: "github.com/sourcegraph/sourcegraph") {
						start
						duration
						command
						dir
						path
					}
				}
			`,
		ExpectedResult: `
				{
					"recordedCommands": []
				}
			`,
	})

	r := rcache.NewFIFOList(wrexec.GetFIFOListKey("github.com/sourcegraph/sourcegraph"), 3)
	err = r.Insert(marshalCmd(t, cmd1))
	require.NoError(t, err)

	RunTest(t, &Test{
		Schema: mustParseGraphQLSchema(t, db),
		Query: `
				{
					recordedCommands(repoName: "github.com/sourcegraph/sourcegraph") {
						start
						duration
						command
						dir
						path
					}
				}
			`,
		ExpectedResult: `
				{
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
					recordedCommands(repoName: "github.com/sourcegraph/sourcegraph") {
						start
						duration
						command
						dir
						path
					}
				}
			`,
		ExpectedResult: `
				{
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
			`,
	})
}

func TestRecordedCommandsResolver_EmptyRepoName(t *testing.T) {
	RunTest(t, &Test{
		Schema: mustParseGraphQLSchema(t, database.NewMockDB()),
		Query: `
				{
					recordedCommands(repoName: "  ") {
						start
						duration
						command
						dir
						path
					}
				}
			`,
		ExpectedResult: "null",
		ExpectedErrors: []*errors.QueryError{
			{
				Message: "empty string provided as repository name",
				Path:    []any{"recordedCommands"},
			},
		},
	})
}

func marshalCmd(t *testing.T, command wrexec.RecordedCommand) []byte {
	t.Helper()
	bytes, err := json.Marshal(&command)
	require.NoError(t, err)
	return bytes
}

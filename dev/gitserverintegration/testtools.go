// package gitserverintegration provides utilities for testing against a real gitserver
// in integration testing.
package gitserverintegration

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/shared"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	v1 "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewTestGitserverWithRepos spawns a new gitserver with the given repos cloned,
// the map holds the repo name to path on disk mappings. The repos will be cloned
// from the location on disk into gitserver, for a most realistic setup.
// Two clients will be returned to interact with the gitserver.
func NewTestGitserverWithRepos(t *testing.T, repos map[api.RepoName]string) (gitserver.Client, v1.GitserverRepositoryServiceClient) {
	// Create supporting infrastructure:
	ctx := context.Background()
	logger := logtest.Scoped(t)
	obsCtx := observation.TestContextTB(t)
	reposDir := t.TempDir()

	// Create a mock database:
	db := dbmocks.NewMockDB()
	repoStore := dbmocks.NewMockRepoStore()
	repoStore.GetByNameFunc.SetDefaultHook(func(ctx context.Context, rn api.RepoName) (*types.Repo, error) {
		if _, ok := repos[rn]; !ok {
			return nil, errors.New("repo not found")
		}
		return &types.Repo{ID: 1, Name: rn}, nil
	})
	db.ReposFunc.SetDefaultReturn(repoStore)
	db.GitserverReposFunc.SetDefaultReturn(dbmocks.NewMockGitserverRepoStore())

	// Spawn a test gitserver on a pseudo random port:
	testAddr := "127.0.0.1:29484"
	routine, err := shared.TestAPIServer(ctx, obsCtx, db, &shared.Config{
		ReposDir:                        reposDir,
		ExhaustiveRequestLoggingEnabled: true,
		ListenAddress:                   testAddr,
	}, func(ctx context.Context, repo api.RepoName) (string, error) {
		if _, ok := repos[repo]; !ok {
			return "", errors.New("invalid repo name passed to getRemoteURL func")
		}

		// We make gitserver clone the repo from the local dir where we create our test repo:
		return repos[repo], nil
	})
	require.NoError(t, err)

	// Start the gitserver up and make sure we shut down cleanly on test exit:
	go routine.Start()
	t.Cleanup(func() {
		require.NoError(t, routine.Stop(context.Background()))
	})

	// Create a gitserver.Client to talk to the gitserver:
	gs := gitserver.NewTestClient(t).WithClientSource(gitserver.NewTestClientSource(t, []string{testAddr}, func(o *gitserver.TestClientSourceOptions) {
		o.ClientFunc = func(conn *grpc.ClientConn) v1.GitserverServiceClient {
			return v1.NewGitserverServiceClient(conn)
		}
	}))

	// Also create a GitserverRepositoryServiceClient to talk to the gitserver:
	conn, err := defaults.Dial(testAddr, logger)
	require.NoError(t, err)
	rs := v1.NewGitserverRepositoryServiceClient(conn)

	// Ensure all the requested repos are cloned into the gitserver:
	for repo := range repos {
		_, err = rs.FetchRepository(ctx, &v1.FetchRepositoryRequest{
			RepoName: string(repo),
		})
		require.NoError(t, err)
	}

	return gs, rs
}

// RepoWithCommands is a helper method to create a git repo with the given commands.
// The repo will be created in a temporary directory, and can be passed to NewTestGitserverWithRepos.
func RepoWithCommands(t *testing.T, cmds ...string) string {
	tmpDir := t.TempDir()
	// Prepare repo state:
	for _, cmd := range append(
		append([]string{"git init --initial-branch=master ."}, cmds...),
		// Promote the repo to a bare repo.
		"git config --bool core.bare true",
	) {
		out, err := gitserver.CreateGitCommand(tmpDir, "bash", "-c", cmd).CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to run git command %v. Output was:\n\n%s", cmd, out)
		}
	}

	return tmpDir
}

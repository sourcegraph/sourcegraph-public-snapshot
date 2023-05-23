package perforce

import (
	"context"
	"database/sql"
	"fmt"
	"path"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/stretchr/testify/require"
)

// setupTestRepo will setup a git repo with 5 commits using p4-fusion as the format in the commit
// messages and returns the directory where the repo is created and a list of (commits, changelist
// IDs) ordered latest to oldest.
func setupTestRepo(t *testing.T) (server.GitDir, []types.PerforceChangelist) {
	commitMessage := `%d - test change

[p4-fusion: depot-paths = "//test-perms/": change = %d]`

	commitCommand := "GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com git commit --allow-empty -m '%s'"

	gitCommands := []string{}
	for cid := 1; cid < 6; cid++ {
		gitCommands = append(gitCommands, fmt.Sprintf(
			commitCommand,
			fmt.Sprintf(commitMessage, cid, cid),
		))
	}

	dir := gitserver.InitGitRepository(t, gitCommands...)

	// Get a list of the commits.
	cmd := gitserver.CreateGitCommand(dir, "bash", "-c", "git rev-list HEAD")
	revList, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run git rev-list HEAD: %q", err.Error())
	}

	commitSHAs := strings.Split(string(revList), "\n")
	allCommitMaps := []types.PerforceChangelist{}

	// Latest commit first, so will have the highest changelist ID (5) and decreases so on until the
	// first commit's changelist ID is 1.
	cid := int64(5)
	for _, commitSHA := range commitSHAs {
		// Drop the last empty item because we split by newline above.
		if commitSHA == "" {
			continue
		}

		allCommitMaps = append(allCommitMaps, types.PerforceChangelist{
			CommitSHA:    api.CommitID(strings.TrimSpace(commitSHA)),
			ChangelistID: cid,
		})

		cid -= 1
	}

	return GitDir(path.Join(dir, ".git")), allCommitMaps
}

func TestGetCommitsToInsert(t *testing.T) {
	dir, allCommitMaps := setupTestRepo(t)

	ctx := context.Background()
	logger := logtest.NoOp(t)
	db := database.NewMockDB()
	repoCommitsStore := database.NewMockRepoCommitsChangelistsStore()
	db.RepoCommitsChangelistsFunc.SetDefaultReturn(repoCommitsStore)

	s := &Server{
		Logger:         logger,
		DB:             db,
		ObservationCtx: observation.TestContextTB(t),
	}

	t.Run("new repo, never mapped", func(t *testing.T) {
		repoCommitsStore.GetLatestForRepoFunc.SetDefaultReturn(nil, sql.ErrNoRows)

		commitMaps, err := s.getCommitsToInsert(ctx, logger, api.RepoID(1), dir)
		require.NoError(t, err)

		if diff := cmp.Diff(allCommitMaps, commitMaps); diff != "" {
			t.Fatalf("mismatched commit maps, (-want,+got)\n:%v", diff)
		}
	})

	t.Run("existing repo, partially mapped", func(t *testing.T) {
		// Commits are latest to oldest and we have a total of 5 commits.
		secondCommit := allCommitMaps[3]

		latestRepoCommit := &types.RepoCommit{
			ID:                   2,
			RepoID:               1,
			CommitSHA:            dbutil.CommitBytea(strings.TrimSpace(string(secondCommit.CommitSHA))),
			PerforceChangelistID: secondCommit.ChangelistID,
		}

		repoCommitsStore.GetLatestForRepoFunc.SetDefaultReturn(latestRepoCommit, nil)

		commitMaps, err := s.getCommitsToInsert(ctx, logger, api.RepoID(1), dir)
		require.NoError(t, err)

		if diff := cmp.Diff(allCommitMaps[:3], commitMaps); diff != "" {
			t.Fatalf("mismatched commit maps, (-want,+got)\n:%v", diff)
		}
	})

	t.Run("existing repo, fully mapped", func(t *testing.T) {
		// Commits are latest to oldest.
		latestCommit := allCommitMaps[0]

		latestRepoCommit := &types.RepoCommit{
			ID:                   2,
			RepoID:               1,
			CommitSHA:            dbutil.CommitBytea(strings.TrimSpace(string(latestCommit.CommitSHA))),
			PerforceChangelistID: latestCommit.ChangelistID,
		}

		repoCommitsStore.GetLatestForRepoFunc.SetDefaultReturn(latestRepoCommit, nil)

		commitMaps, err := s.getCommitsToInsert(ctx, logger, api.RepoID(1), dir)
		require.NoError(t, err)
		require.Nil(t, commitMaps)
	})
}

func TestHeadCommitSHA(t *testing.T) {
	dir, allCommitMaps := setupTestRepo(t)

	ctx := context.Background()
	logger := logtest.NoOp(t)

	commitSHA, err := headCommitSHA(ctx, logger, dir)

	require.NoError(t, err)
	require.Equal(t, string(allCommitMaps[0].CommitSHA), commitSHA)
}

func TestNewMappableCommits(t *testing.T) {
	ctx := context.Background()
	logger := logtest.NoOp(t)

	dir, allCommitMaps := setupTestRepo(t)

	gotCommitMaps, err := newMappableCommits(ctx, logger, dir, "", "")
	require.NoError(t, err, "unexpected error in newMapppableCommits")

	if diff := cmp.Diff(allCommitMaps, gotCommitMaps); diff != "" {
		t.Fatalf("mismatched commit maps, (-want,+got)\n:%v", diff)
	}
}

func TestReadGitLogOutput(t *testing.T) {
	logger := logtest.NoOp(t)

	t.Run("clean exit, no error", func(t *testing.T) {
		ctx := context.Background()

		logLineResults := make(chan string)
		reader := strings.NewReader("line1\nline2\nline3\n")

		go func() {
			err := readGitLogOutput(ctx, logger, reader, logLineResults)
			if err != nil {
				panic(fmt.Sprintf("unexpected error from readGitLogOutput: %q", err.Error()))
			}
		}()

		// Check that we received all gotLines before the timeout
		var gotLines []string
		for i := 0; i < 3; i++ {
			l := <-logLineResults
			gotLines = append(gotLines, l)
		}

		// If readGitLogOutput attempts to write more than the three expected items, this will cause a
		// panic with "send on closed channel".
		close(logLineResults)

		wantLines := []string{"line1", "line2", "line3"}

		if diff := cmp.Diff(wantLines, gotLines); diff != "" {
			t.Fatalf("mismatched lines, (-want,+got)\n:%v", diff)
		}

		require.NoError(t, ctx.Err(), "unexpected error in context")
	})

	t.Run("early exit, with error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		logLineResults := make(chan string)
		reader := strings.NewReader("line1\nline2\nline3\n")

		go func() {
			err := readGitLogOutput(ctx, logger, reader, logLineResults)
			if err != nil {
				panic(fmt.Sprintf("unexpected error from readGitLogOutput: %q", err.Error()))
			}
		}()

		// Cancel the context without reading anything to trigger an error. This should cause the
		// goroutine to exit early and return the error.
		cancel()

		require.NoError(t, ctx.Err(), "no error in context")
	})
}

func TestParseGitLogLint(t *testing.T) {
	t.Run("passes valid perforce commit", func(t *testing.T) {
		got, err := parseGitLogLine(`4e5b9dbc6393b195688a93ea04b98fada50bfa03 [p4-fusion: depot-paths = "//rhia-depot-test/": change = 83733]`)

		want := &types.PerforceChangelist{
			CommitSHA:    api.CommitID("4e5b9dbc6393b195688a93ea04b98fada50bfa03"),
			ChangelistID: 83733,
		}

		require.NoError(t, err)
		require.Equal(t, want, got)
	})

	t.Run("fails invalid perforce commit", func(t *testing.T) {
		got, err := parseGitLogLine(`4e5b9dbc6393b195688a93ea04b98fada50bfa03 invalid format`)

		require.Error(t, err)
		require.Nil(t, got)
	})
}

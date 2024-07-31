package rockskip

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search"
)

func TestIndex(t *testing.T) {
	repo, repoDir := gitserver.MakeGitRepositoryAndReturnDir(t)
	// Needed in CI
	gitRun(t, repoDir, "config", "user.email", "test@sourcegraph.com")

	git, err := newSubprocessGit(t, repoDir)
	require.NoError(t, err)
	defer git.Close()

	db, service := mockService(t, git)
	defer db.Close()

	state := map[string][]string{}
	verifyBlobs := func(lang string, ext string) {
		out, err := gitserver.CreateGitCommand(repoDir, "git", "rev-parse", "HEAD").CombinedOutput()
		require.NoError(t, err, string(out))
		commit := string(bytes.TrimSpace(out))

		args := search.SymbolsParameters{
			Repo:         repo,
			CommitID:     api.CommitID(commit),
			Query:        "",
			IncludeLangs: []string{lang}}
		symbols, err := service.Search(context.Background(), args)
		require.NoError(t, err)

		// Make sure the paths match.
		gotPathSet := map[string]struct{}{}
		for _, blob := range symbols {
			gotPathSet[blob.Path] = struct{}{}
		}
		gotPaths := []string{}
		for gotPath := range gotPathSet {
			gotPaths = append(gotPaths, gotPath)
		}
		wantPaths := []string{}
		for wantPath := range state {
			if strings.Contains(wantPath, ext) {
				wantPaths = append(wantPaths, wantPath)
			}
		}
		sort.Strings(gotPaths)
		sort.Strings(wantPaths)
		if diff := cmp.Diff(gotPaths, wantPaths); diff != "" {
			fmt.Println("unexpected paths (-got +want)")
			fmt.Println(diff)
			err = PrintInternals(context.Background(), db)
			require.NoError(t, err)
			t.FailNow()
		}

		gotPathToSymbols := map[string][]string{}
		for _, blob := range symbols {
			gotPathToSymbols[blob.Path] = append(gotPathToSymbols[blob.Path], blob.Name)
		}

		// Make sure the symbols match.
		for gotPath, gotSymbols := range gotPathToSymbols {
			wantSymbols := state[gotPath]
			sort.Strings(gotSymbols)
			sort.Strings(wantSymbols)
			if diff := cmp.Diff(gotSymbols, wantSymbols); diff != "" {
				fmt.Println("unexpected symbols (-got +want)")
				fmt.Println(diff)
				err = PrintInternals(context.Background(), db)
				require.NoError(t, err)
				t.FailNow()
			}
		}
	}

	gitAdd(t, repoDir, state, "a.txt", "sym1\n")
	verifyBlobs("Text", ".txt")

	gitAdd(t, repoDir, state, "b.txt", "sym1\n")
	verifyBlobs("Text", ".txt")

	gitAdd(t, repoDir, state, "c.txt", "sym1\nsym2")
	verifyBlobs("Text", ".txt")

	gitAdd(t, repoDir, state, "a.java", "sym1\nsym2")
	verifyBlobs("Java", ".java")

	gitAdd(t, repoDir, state, "a.txt", "sym1\nsym2")
	verifyBlobs("Text", ".txt")

	gitRm(t, repoDir, state, "a.txt")
	verifyBlobs("Text", ".txt")
}

func TestRuler(t *testing.T) {
	testCases := []struct {
		n    int
		want int
	}{
		{0, 0},
		{1, 0},
		{2, 1},
		{3, 0},
		{4, 2},
		{5, 0},
		{6, 1},
		{7, 0},
		{8, 3},
		{64, 6},
		{96, 5},
		{123, 0},
	}

	for _, tc := range testCases {
		got := ruler(tc.n)
		if got != tc.want {
			t.Errorf("ruler(%d): got %d, want %d", tc.n, got, tc.want)
		}
	}
}

func TestGetHops(t *testing.T) {
	ctx := context.Background()
	repoId := 42

	db := dbtest.NewDB(t)
	defer db.Close()

	// Insert some initial commits.
	commit0, err := InsertCommit(ctx, db, repoId, "0000", 0, NULL)
	require.NoError(t, err)

	commit1, err := InsertCommit(ctx, db, repoId, "1111", 0, commit0)
	require.NoError(t, err)

	commit2, err := InsertCommit(ctx, db, repoId, "2222", 1, commit0)
	require.NoError(t, err)

	commit3, err := InsertCommit(ctx, db, repoId, "3333", 0, commit2)
	require.NoError(t, err)

	commit4, err := InsertCommit(ctx, db, repoId, "4444", 2, NULL)
	require.NoError(t, err)

	tests := []struct {
		name   string
		commit CommitId
		want   []int
	}{
		{
			name:   "commit0",
			commit: commit0,
			want:   []CommitId{commit0, NULL},
		},
		{
			name:   "commit1",
			commit: commit1,
			want:   []CommitId{commit1, commit0, NULL},
		},
		{
			name:   "commit2",
			commit: commit2,
			want:   []CommitId{commit2, commit0, NULL},
		},
		{
			name:   "commit3",
			commit: commit3,
			want:   []CommitId{commit3, commit2, commit0, NULL},
		},
		{
			name:   "commit4",
			commit: commit4,
			want:   []CommitId{commit4, NULL},
		},
		{
			name:   "nonexistent",
			commit: 42,
			want:   []CommitId{42},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getHops(ctx, db, tt.commit, NewTaskLog())
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

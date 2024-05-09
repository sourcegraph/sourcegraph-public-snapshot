package rockskip

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/go-ctags"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/fetcher"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// mockParser converts each line to a symbol.
type mockParser struct{}

func (mockParser) Parse(path string, bytes []byte) ([]*ctags.Entry, error) {
	symbols := []*ctags.Entry{}

	for lineNumber, line := range strings.Split(string(bytes), "\n") {
		if line == "" {
			continue
		}

		symbols = append(symbols, &ctags.Entry{Name: line, Line: lineNumber + 1})
	}

	return symbols, nil
}

func (mockParser) Close() {}

func TestIndex(t *testing.T) {
	gitserver.ClientMocks.LocalGitserver = true
	t.Cleanup(gitserver.ResetClientMocks)
	repo, repoDir := gitserver.MakeGitRepositoryAndReturnDir(t)

	state := map[string][]string{}

	gitRun := func(args ...string) {
		out, err := gitserver.CreateGitCommand(repoDir, "git", args...).CombinedOutput()
		require.NoError(t, err, string(out))
	}

	add := func(filename string, contents string) {
		require.NoError(t, os.WriteFile(path.Join(repoDir, filename), []byte(contents), 0644), "os.WriteFile")
		gitRun("add", filename)
		symbols, err := mockParser{}.Parse(filename, []byte(contents))
		require.NoError(t, err)
		state[filename] = []string{}
		for _, symbol := range symbols {
			state[filename] = append(state[filename], symbol.Name)
		}
	}

	rm := func(filename string) {
		gitRun("rm", filename)
		delete(state, filename)
	}

	gitRun("init")
	// Needed in CI
	gitRun("config", "user.email", "test@sourcegraph.com")

	git, err := NewSubprocessGit(t, repoDir)
	require.NoError(t, err)
	defer git.Close()

	db := dbtest.NewDB(t)
	defer db.Close()

	createParser := func() (ctags.Parser, error) { return mockParser{}, nil }

	service, err := NewService(observation.TestContextTB(t), db, git, newMockRepositoryFetcher(git), createParser, 1, 1, false, 1, 1, 1, false)
	require.NoError(t, err)

	verifyBlobs := func() {
		out, err := gitserver.CreateGitCommand(repoDir, "git", "rev-parse", "HEAD").CombinedOutput()
		require.NoError(t, err, string(out))
		commit := string(bytes.TrimSpace(out))

		args := search.SymbolsParameters{
			Repo:         api.RepoName(repo),
			CommitID:     api.CommitID(commit),
			Query:        "",
			IncludeLangs: []string{"Text"}}
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
			// We only want .txt files since we're filtering by lang: text
			if strings.Contains(wantPath, ".txt") {
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

	commit := func(message string) {
		gitRun("commit", "--allow-empty", "-m", message)
		verifyBlobs()
	}

	add("a.txt", "sym1\n")
	commit("add a file with 1 symbol")

	add("b.txt", "sym1\n")
	commit("add another file with 1 symbol")

	add("c.txt", "sym1\nsym2")
	commit("add another file with 2 symbols")

	add("a.java", "sym1\nsym2")
	commit("System.out.println(\"hello, world!\"")

	add("a.txt", "sym1\nsym2")
	commit("add a symbol to a.txt")

	commit("empty")

	rm("a.txt")
	commit("rm a.txt")
}

type SubprocessGit struct {
	gitDir        string
	catFileCmd    *exec.Cmd
	catFileStdin  io.WriteCloser
	catFileStdout bufio.Reader
	gs            gitserver.Client
}

func NewSubprocessGit(t testing.TB, gitDir string) (*SubprocessGit, error) {
	cmd := exec.Command("git", "cat-file", "--batch")
	cmd.Dir = gitDir

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	return &SubprocessGit{
		gitDir:        gitDir,
		catFileCmd:    cmd,
		catFileStdin:  stdin,
		catFileStdout: *bufio.NewReader(stdout),
		gs:            gitserver.NewTestClient(t),
	}, nil
}

func (g SubprocessGit) Close() error {
	err := g.catFileStdin.Close()
	if err != nil {
		return err
	}
	return g.catFileCmd.Wait()
}

func (g SubprocessGit) LogReverseEach(ctx context.Context, repo string, givenCommit string, n int, onLogEntry func(entry gitdomain.LogEntry) error) (returnError error) {
	log := exec.Command("git", gitdomain.LogReverseArgs(n, givenCommit)...)
	log.Dir = g.gitDir
	output, err := log.StdoutPipe()
	if err != nil {
		return err
	}

	err = log.Start()
	if err != nil {
		return err
	}
	defer func() {
		err = log.Wait()
		if err != nil {
			returnError = err
		}
	}()

	return gitdomain.ParseLogReverseEach(output, onLogEntry)
}

func (g *SubprocessGit) RevList(ctx context.Context, repo string, commit string, onCommit func(commit string) (shouldContinue bool, err error)) (returnError error) {
	nextCursor := commit
	for {
		var commits []api.CommitID
		var err error
		commits, nextCursor, err = g.paginatedRevList(ctx, api.RepoName(repo), nextCursor, 100)
		if err != nil {
			return err
		}
		for _, c := range commits {
			shouldContinue, err := onCommit(string(c))
			if err != nil {
				return err
			}
			if !shouldContinue {
				return nil
			}
		}
		if nextCursor == "" {
			return nil
		}
	}
}

func (g *SubprocessGit) paginatedRevList(ctx context.Context, repo api.RepoName, commit string, count int) (_ []api.CommitID, nextCursor string, _ error) {
	commits, err := g.gs.Commits(ctx, repo, gitserver.CommitsOptions{
		N:           uint(count + 1),
		Range:       commit,
		FirstParent: true,
	})
	if err != nil {
		return nil, "", err
	}

	commitIDs := make([]api.CommitID, 0, count+1)

	for _, commit := range commits {
		commitIDs = append(commitIDs, commit.ID)
	}

	if len(commitIDs) > count {
		nextCursor = string(commitIDs[len(commitIDs)-1])
		commitIDs = commitIDs[:count]
	}

	return commitIDs, nextCursor, nil
}

func newMockRepositoryFetcher(git *SubprocessGit) fetcher.RepositoryFetcher {
	return &mockRepositoryFetcher{git: git}
}

type mockRepositoryFetcher struct{ git *SubprocessGit }

func (f *mockRepositoryFetcher) FetchRepositoryArchive(ctx context.Context, repo api.RepoName, commit api.CommitID, paths []string) <-chan fetcher.ParseRequestOrError {
	ch := make(chan fetcher.ParseRequestOrError)

	go func() {
		for _, p := range paths {
			_, err := f.git.catFileStdin.Write([]byte(fmt.Sprintf("%s:%s\n", commit, p)))
			if err != nil {
				ch <- fetcher.ParseRequestOrError{
					Err: errors.Wrap(err, "writing to cat-file stdin"),
				}
				return
			}

			line, err := f.git.catFileStdout.ReadString('\n')
			if err != nil {
				ch <- fetcher.ParseRequestOrError{
					Err: errors.Wrap(err, "read newline"),
				}
				return
			}
			line = line[:len(line)-1] // Drop the trailing newline
			parts := strings.Split(line, " ")
			if len(parts) != 3 {
				ch <- fetcher.ParseRequestOrError{
					Err: errors.Newf("unexpected cat-file output: %q", line),
				}
				return
			}
			size, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil {
				ch <- fetcher.ParseRequestOrError{
					Err: errors.Wrap(err, "parse size"),
				}
				return
			}

			fileContents, err := io.ReadAll(io.LimitReader(&f.git.catFileStdout, size))
			if err != nil {
				ch <- fetcher.ParseRequestOrError{
					Err: errors.Wrap(err, "read contents"),
				}
				return
			}

			discarded, err := f.git.catFileStdout.Discard(1) // Discard the trailing newline
			if err != nil {
				ch <- fetcher.ParseRequestOrError{
					Err: errors.Wrap(err, "discard newline"),
				}
				return
			}
			if discarded != 1 {
				ch <- fetcher.ParseRequestOrError{
					Err: errors.Newf("expected to discard 1 byte, but discarded %d", discarded),
				}
				return
			}

			ch <- fetcher.ParseRequestOrError{
				ParseRequest: fetcher.ParseRequest{
					Path: p,
					Data: fileContents,
				},
			}
		}

		close(ch)
	}()

	return ch
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

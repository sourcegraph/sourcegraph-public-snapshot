package rockskip

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/sourcegraph/go-ctags"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/fetcher"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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

func mockService(t *testing.T, git *subprocessGit) (*sql.DB, *Service) {
	observationCtx := &observation.TestContext
	db := dbtest.NewDB(t)
	repoFetcher := newMockRepositoryFetcher(git)
	createParser := func() (ctags.Parser, error) { return mockParser{}, nil }

	service, err := NewService(observationCtx, db, git, repoFetcher, createParser, 1, 1, false, 1, 1, 1, false)
	require.NoError(t, err)

	return db, service
}

func gitRm(t *testing.T, repoDir string, state map[string][]string, filename string) {
	gitRun(t, repoDir, "rm", filename)
	delete(state, filename)
	gitRun(t, repoDir, "commit", "--allow-empty", "-m", "remove "+filename)
}

func gitAdd(t *testing.T, repoDir string, state map[string][]string, filename string, contents string) {
	require.NoError(t, os.WriteFile(path.Join(repoDir, filename), []byte(contents), 0644), "os.WriteFile")
	gitRun(t, repoDir, "add", filename)
	symbols, err := mockParser{}.Parse(filename, []byte(contents))
	require.NoError(t, err)
	state[filename] = []string{}
	for _, symbol := range symbols {
		state[filename] = append(state[filename], symbol.Name)
	}
	gitRun(t, repoDir, "commit", "--allow-empty", "-m", "add "+filename)
}

func gitRun(t *testing.T, repoDir string, args ...string) {
	out, err := gitserver.CreateGitCommand(repoDir, "git", args...).CombinedOutput()
	require.NoError(t, err, string(out))
}

type subprocessGit struct {
	gitDir        string
	catFileCmd    *exec.Cmd
	catFileStdin  io.WriteCloser
	catFileStdout bufio.Reader
	gs            gitserver.Client
}

func newSubprocessGit(t testing.TB, gitDir string) (*subprocessGit, error) {
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

	return &subprocessGit{
		gitDir:        gitDir,
		catFileCmd:    cmd,
		catFileStdin:  stdin,
		catFileStdout: *bufio.NewReader(stdout),
		gs:            gitserver.NewTestClient(t),
	}, nil
}

func (g subprocessGit) Close() error {
	err := g.catFileStdin.Close()
	if err != nil {
		return err
	}
	return g.catFileCmd.Wait()
}

func (g subprocessGit) LogReverseEach(_ context.Context, _ string, givenCommit string, n int, onLogEntry func(entry gitdomain.LogEntry) error) (returnError error) {
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

func (g *subprocessGit) RevList(ctx context.Context, repo string, commit string, onCommit func(commit string) (shouldContinue bool, err error)) (returnError error) {
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

func (g *subprocessGit) paginatedRevList(ctx context.Context, repo api.RepoName, commit string, count int) (_ []api.CommitID, nextCursor string, _ error) {
	commits, err := g.gs.Commits(ctx, repo, gitserver.CommitsOptions{
		N:           uint(count + 1),
		Ranges:      []string{commit},
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

func newMockRepositoryFetcher(git *subprocessGit) fetcher.RepositoryFetcher {
	return &mockRepositoryFetcher{git: git}
}

type mockRepositoryFetcher struct{ git *subprocessGit }

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

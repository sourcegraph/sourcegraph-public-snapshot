package rockskip

import (
	"bufio"
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
	"github.com/sourcegraph/sourcegraph/cmd/symbols/fetcher"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
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
	fatalIfError := func(err error, message string) {
		if err != nil {
			t.Fatal(errors.Wrap(err, message))
		}
	}

	gitDir, err := os.MkdirTemp("", "rockskip-test-index")
	fatalIfError(err, "faiMkdirTemp")

	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("git dir %s left intact for inspection", gitDir)
		} else {
			os.RemoveAll(gitDir)
		}
	})

	gitCmd := func(args ...string) *exec.Cmd {
		cmd := exec.Command("git", args...)
		cmd.Dir = gitDir
		return cmd
	}

	gitRun := func(args ...string) {
		fatalIfError(gitCmd(args...).Run(), "git "+strings.Join(args, " "))
	}

	gitStdout := func(args ...string) string {
		stdout, err := gitCmd(args...).Output()
		fatalIfError(err, "git "+strings.Join(args, " "))
		return string(stdout)
	}

	getHead := func() string {
		return strings.TrimSpace(gitStdout("rev-parse", "HEAD"))
	}

	state := map[string][]string{}

	add := func(filename string, contents string) {
		fatalIfError(os.WriteFile(path.Join(gitDir, filename), []byte(contents), 0644), "os.WriteFile")
		gitRun("add", filename)
		symbols, err := mockParser{}.Parse(filename, []byte(contents))
		fatalIfError(err, "simpleParse")
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

	git, err := NewSubprocessGit(gitDir)
	fatalIfError(err, "NewSubprocessGit")
	defer git.Close()

	db := dbtest.NewDB(t)
	defer db.Close()

	createParser := func() (ctags.Parser, error) { return mockParser{}, nil }

	service, err := NewService(db, git, newMockRepositoryFetcher(git), createParser, 1, 1, false, 1, 1, 1, false)
	fatalIfError(err, "NewService")

	verifyBlobs := func() {
		repo := "somerepo"
		commit := getHead()
		args := search.SymbolsParameters{Repo: api.RepoName(repo), CommitID: api.CommitID(commit), Query: ""}
		symbols, err := service.Search(context.Background(), args)
		fatalIfError(err, "Search")

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
			wantPaths = append(wantPaths, wantPath)
		}
		sort.Strings(gotPaths)
		sort.Strings(wantPaths)
		if diff := cmp.Diff(gotPaths, wantPaths); diff != "" {
			fmt.Println("unexpected paths (-got +want)")
			fmt.Println(diff)
			err = PrintInternals(context.Background(), db)
			fatalIfError(err, "PrintInternals")
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
				fatalIfError(err, "PrintInternals")
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
}

func NewSubprocessGit(gitDir string) (*SubprocessGit, error) {
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

func (g SubprocessGit) RevList(ctx context.Context, repo string, givenCommit string, onCommit func(commit string) (shouldContinue bool, err error)) (returnError error) {
	revList := exec.Command("git", gitserver.RevListArgs(givenCommit)...)
	revList.Dir = g.gitDir
	output, err := revList.StdoutPipe()
	if err != nil {
		return err
	}

	err = revList.Start()
	if err != nil {
		return err
	}
	defer func() {
		err = revList.Wait()
		if err != nil {
			returnError = err
		}
	}()

	return gitdomain.RevListEach(output, onCommit)
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

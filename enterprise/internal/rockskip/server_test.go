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

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// simpleParse converts each line into a symbol.
func simpleParse(path string, bytes []byte) ([]Symbol, error) {
	symbols := []Symbol{}

	for _, line := range strings.Split(string(bytes), "\n") {
		if line == "" {
			continue
		}

		symbols = append(symbols, Symbol{Name: line})
	}

	return symbols, nil
}

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
		symbols, err := simpleParse(filename, []byte(contents))
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

	git, err := NewSubprocessGit(gitDir)
	fatalIfError(err, "NewSubprocessGit")
	defer git.Close()

	db := dbtest.NewDB(t)
	defer db.Close()

	createParser := func() ParseSymbolsFunc { return simpleParse }

	service, err := NewService(db, git, createParser, 1, 1, false, 1, 1, 1)
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
		for path := range gotPathSet {
			gotPaths = append(gotPaths, path)
		}
		wantPaths := []string{}
		for path := range state {
			wantPaths = append(wantPaths, path)
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
		for path, gotSymbols := range gotPathToSymbols {
			wantSymbols := state[path]
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

func (g SubprocessGit) LogReverseEach(repo string, db database.DB, givenCommit string, n int, onLogEntry func(entry gitdomain.LogEntry) error) (returnError error) {
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

func (g SubprocessGit) RevListEach(repo string, db database.DB, givenCommit string, onCommit func(commit string) (shouldContinue bool, err error)) (returnError error) {
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

	return gitserver.NewClient(db).RevListEach(output, onCommit)
}

func (g SubprocessGit) ArchiveEach(repo string, commit string, paths []string, onFile func(path string, contents []byte) error) error {
	for _, path := range paths {
		_, err := g.catFileStdin.Write([]byte(fmt.Sprintf("%s:%s\n", commit, path)))
		if err != nil {
			return errors.Wrap(err, "writing to cat-file stdin")
		}

		line, err := g.catFileStdout.ReadString('\n')
		if err != nil {
			return errors.Wrap(err, "read newline")
		}
		line = line[:len(line)-1] // Drop the trailing newline
		parts := strings.Split(line, " ")
		if len(parts) != 3 {
			return errors.Newf("unexpected cat-file output: %q", line)
		}
		size, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return errors.Wrap(err, "parse size")
		}

		fileContents, err := io.ReadAll(io.LimitReader(&g.catFileStdout, size))
		if err != nil {
			return errors.Wrap(err, "read contents")
		}

		discarded, err := g.catFileStdout.Discard(1) // Discard the trailing newline
		if err != nil {
			return errors.Wrap(err, "discard newline")
		}
		if discarded != 1 {
			return errors.Newf("expected to discard 1 byte, but discarded %d", discarded)
		}

		err = onFile(path, fileContents)
		if err != nil {
			return errors.Wrap(err, "onFile")
		}
	}

	return nil
}

package rockskip

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/sync/semaphore"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
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
		fatalIfError(ioutil.WriteFile(path.Join(gitDir, filename), []byte(contents), 0644), "ioutil.WriteFile")
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

	verifyBlobs := func() {
		repo := "somerepo"
		commit := getHead()
		status := NewRequestStatus(repo, commit, func() {})
		args := types.SearchArgs{Repo: api.RepoName(repo), CommitID: api.CommitID(commit), Query: ""}
		blobs, err := Search(context.Background(), args, git, db, simpleParse, 1, semaphore.NewWeighted(1), semaphore.NewWeighted(1), status, make(chan<- struct{}))
		fatalIfError(err, "Search")

		// Make sure the paths match.
		gotPaths := []string{}
		for _, blob := range blobs {
			gotPaths = append(gotPaths, blob.Path)
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
			PrintInternals(context.Background(), db)
			t.FailNow()
		}

		// Make sure the symbols match.
		for _, blob := range blobs {
			gotSymbols := []string{}
			for _, symbol := range blob.Symbols {
				gotSymbols = append(gotSymbols, symbol.Name)
			}
			wantSymbols := state[blob.Path]
			sort.Strings(gotSymbols)
			sort.Strings(wantSymbols)
			if diff := cmp.Diff(gotSymbols, wantSymbols); diff != "" {
				fmt.Println("unexpected symbols (-got +want)")
				fmt.Println(diff)
				PrintInternals(context.Background(), db)
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

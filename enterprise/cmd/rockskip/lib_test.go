package rockskip

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/go-ctags"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

type Ctags struct {
	parser ctags.Parser
}

func NewCtags() (Ctags, error) {
	parser, err := ctags.New(ctags.Options{
		Bin:                "ctags",
		PatternLengthLimit: 0,
	})
	if err != nil {
		return Ctags{}, err
	}
	return Ctags{
		parser: parser,
	}, nil
}

func (ctags Ctags) Parse(path string, bytes []byte) (symbols []Symbol, err error) {
	symbols = []Symbol{}
	entries, err := ctags.parser.Parse(path, bytes)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		symbols = append(symbols, Symbol{
			Name:   entry.Name,
			Parent: entry.Parent,
			Kind:   entry.Kind,
			Line:   entry.Line,
		})
	}
	return symbols, nil
}

func (ctags Ctags) Close() {
	ctags.parser.Close()
}

func TestIndex(t *testing.T) {
	repo := "github.com/gorilla/mux"
	// repo := "github.com/hashicorp/raft"
	// repo := "github.com/crossplane/crossplane"
	// repo := "github.com/kubernetes/kubernetes"

	git, err := NewSubprocessGit(repo)
	if err != nil {
		t.Fatalf("🚨 NewSubprocessGit: %s", err)
	}
	defer git.Close()

	db := dbtest.NewDB(t)

	defer db.Close()
	parser, err := NewCtags()
	if err != nil {
		t.Fatalf("🚨 NewCtags: %s", err)
	}
	defer parser.Close()

	revParse := exec.Command("git", "rev-parse", "HEAD")
	revParse.Dir = "/Users/chrismwendt/" + repo
	output, err := revParse.Output()
	if err != nil {
		t.Fatalf("🚨 rev-parse: %s", err)
	}
	head := strings.TrimSpace(string(output))

	TASKLOG.Reset()
	err = Index(git, db, parser.Parse, head)
	if err != nil {
		t.Fatalf("🚨 Index: %s", err)
	}
	TASKLOG.Print()

	blobs, err := Search(db, head, nil)
	if err != nil {
		t.Fatalf("🚨 PathsAtCommit: %s", err)
	}
	paths := []string{}
	for _, blob := range blobs {
		paths = append(paths, blob.Path)
	}

	cmd := exec.Command("bash", "-c", fmt.Sprintf("git ls-tree -r %s | grep -v \"^160000\" | cut -f2", head))
	cmd.Dir = "/Users/chrismwendt/" + repo
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	expected := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")

	sort.Strings(paths)
	sort.Strings(expected)

	if diff := cmp.Diff(paths, expected); diff != "" {
		fmt.Println("🚨 PathsAtCommit: unexpected paths (-got +want)")
		fmt.Println(diff)
		PrintInternals(db)
		t.Fail()
	}
}

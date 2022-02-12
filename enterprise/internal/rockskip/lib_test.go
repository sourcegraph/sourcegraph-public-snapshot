package rockskip

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/go-ctags"
	"golang.org/x/sync/semaphore"

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

const HOME = "/Users/chrismwendt/"

func TestIndex(t *testing.T) {
	// repo := "github.com/gorilla/mux"
	// repo := "github.com/hashicorp/raft"
	// repo := "github.com/crossplane/crossplane"
	// repo := "github.com/kubernetes/kubernetes"
	repo := "github.com/hashicorp/go-multierror"

	git, err := NewSubprocessGit(repo)
	if err != nil {
		t.Fatalf("🚨 NewSubprocessGit: %s", err)
	}
	defer git.Close()

	db := dbtest.NewDB(t)
	fmt.Println()
	defer db.Close()

	parser, err := NewCtags()
	if err != nil {
		t.Fatalf("🚨 NewCtags: %s", err)
	}
	defer parser.Close()

	revParse := exec.Command("git", "rev-parse", "HEAD~1")
	revParse.Dir = HOME + repo
	output, err := revParse.Output()
	if err != nil {
		t.Fatalf("🚨 rev-parse: %s", err)
	}
	commit := strings.TrimSpace(string(output))

	// du -sh
	du := exec.Command("du", "-sh", HOME+repo)
	du.Dir = HOME + repo
	output, err = du.Output()
	if err != nil {
		t.Fatalf("🚨 du: %s", err)
	}
	size := strings.Split(string(output), "\t")[0]

	fmt.Println("🔵 Indexing", repo, "at", commit, "with git size", size)
	fmt.Println()

	tasklog := NewTaskLog()
	err = Index(git, db, tasklog, parser.Parse, repo, commit, 1, semaphore.NewWeighted(1), NewStatus(repo, commit, tasklog))
	if err != nil {
		t.Fatalf("🚨 Index: %s", err)
	}
	tasklog.Print()
	fmt.Println()

	rows, err := db.Query("SELECT pg_size_pretty(pg_total_relation_size('rockskip_ancestry')) AS rockskip_ancestry, pg_size_pretty(pg_total_relation_size('rockskip_blobs')) AS rockskip_blobs;")
	if err != nil {
		t.Fatalf("🚨 db.Query: %s", err)
	}
	defer rows.Close()
	for rows.Next() {
		var rockskip_ancestry, rockskip_blobs string
		if err := rows.Scan(&rockskip_ancestry, &rockskip_blobs); err != nil {
			t.Fatalf("🚨 rows.Scan: %s", err)
		}
		fmt.Printf("rockskip_ancestry: %s\n", rockskip_ancestry)
		fmt.Printf("rockskip_blobs   : %s\n", rockskip_blobs)
	}
	fmt.Println()

	blobs, err := Search(db, NewTaskLog(), repo, commit, nil)
	if err != nil {
		t.Fatalf("🚨 PathsAtCommit: %s", err)
	}
	paths := []string{}
	for _, blob := range blobs {
		paths = append(paths, blob.Path)
	}

	cmd := exec.Command("bash", "-c", fmt.Sprintf("git ls-tree -r %s | grep -v \"^160000\" | cut -f2", commit))
	cmd.Dir = HOME + repo
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

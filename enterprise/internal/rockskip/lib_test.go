package main

import (
	"fmt"
	"math/rand"
	"os/exec"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/go-ctags"
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

func (ctags Ctags) Parse(path string, bytes []byte) (symbols []string, err error) {
	symbols = []string{}
	entries, err := ctags.parser.Parse(path, bytes)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		symbols = append(symbols, entry.Name)
	}
	return symbols, nil
}

func (ctags Ctags) Close() {
	ctags.parser.Close()
}

func TestIndex(t *testing.T) {
	// repo := "github.com/gorilla/mux"
	// head := "3cf0d013e53d62a96c096366d300c84489c26dd5"
	// repo := "github.com/hashicorp/raft"
	// head := "aa1afe5d2a1e961ef54726af645ede516c18a554"
	repo := "github.com/crossplane/crossplane"
	head := "1f84012248a350b479a575214c17af5fe183138b"

	git, err := NewSubprocessGit(repo)
	if err != nil {
		t.Fatalf("🚨 NewSubprocessGit: %s", err)
	}
	defer git.Close()
	db, err := NewPostgresDB()
	if err != nil {
		t.Fatalf("🚨 NewPostgresDB: %s", err)
	}
	defer db.Close()
	parser, err := NewCtags()
	if err != nil {
		t.Fatalf("🚨 NewCtags: %s", err)
	}
	defer parser.Close()

	INSTANTS.Reset()
	err = Index(git, db, parser.Parse, head)
	if err != nil {
		t.Fatalf("🚨 Index: %s", err)
	}
	INSTANTS.Print()

	blobs, err := Search(db, head)
	if err != nil {
		t.Fatalf("🚨 PathsAtCommit: %s", err)
	}
	paths := []string{}
	for _, blob := range blobs {
		paths = append(paths, blob.path)
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
		db.PrintInternals()
		t.Fail()
	}
}

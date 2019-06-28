package graphqlbackend

import (
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/gituri"
	"github.com/sourcegraph/sourcegraph/pkg/symbols/protocol"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

func TestMakeFileMatchURIFromSymbol(t *testing.T) {
	symbol := protocol.Symbol{
		Name:    "test",
		Path:    "foo/bar",
		Line:    0,
		Pattern: "",
	}
	baseURI, _ := gituri.Parse("https://github.com/foo/bar")
	gitSignatureWithDate := git.Signature{Date: time.Now().UTC().AddDate(0, 0, -1)}
	commit := &gitCommitResolver{
		repo:   &repositoryResolver{repo: &db.MinimalRepo{ID: 1, Name: "repo"}},
		oid:    "c1",
		author: *toSignatureResolver(&gitSignatureWithDate),
	}
	sr := &searchSymbolResult{symbol, baseURI, "go", commit}

	tests := []struct {
		rev  string
		want string
	}{
		{"", "git://repo#foo/bar"},
		{"test", "git://repo?test#foo/bar"},
	}

	for _, test := range tests {
		got := makeFileMatchURIFromSymbol(sr, test.rev)
		if got != test.want {
			t.Errorf("rev(%v) got %v want %v", test.rev, got, test.want)
		}
	}
}

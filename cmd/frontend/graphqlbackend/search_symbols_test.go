package graphqlbackend

import (
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/gituri"
	"github.com/sourcegraph/sourcegraph/pkg/symbols/protocol"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

func TestMakeFileMatchURIFromSymbol(t *testing.T) {
	symbol := protocol.Symbol{
		Name:    "test",
		Path:    "test/path",
		Line:    0,
		Pattern: "",
	}
	baseURI, _ := gituri.Parse("https://github.com/foo/bar?v#f/d")
	gitSignatureWithDate := git.Signature{Date: time.Now().UTC().AddDate(0, 0, -1)}
	commit := &gitCommitResolver{
		repo:   &repositoryResolver{repo: &types.Repo{ID: 1, Name: "repo"}},
		oid:    "c1",
		author: *toSignatureResolver(&gitSignatureWithDate),
	}
	sr := &searchSymbolResult{symbol, baseURI, "go", commit}

	tests := []struct {
		rev  string
		want string
	}{
		{"", "git://repo#f/d"},
		{"test", "git://repo?test#f/d"},
	}

	for _, test := range tests {
		got := makeFileMatchURIFromSymbol(sr, test.rev)
		if got != test.want {
			t.Errorf("rev(%v) got %v want %v", test.rev, got, test.want)
		}
	}
}

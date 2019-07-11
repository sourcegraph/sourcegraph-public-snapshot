package graphqlbackend

import (
	"fmt"
	"math/rand"
	"reflect"
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
		Path:    "foo/bar",
		Line:    0,
		Pattern: "",
	}
	baseURI, _ := gituri.Parse("https://github.com/foo/bar")
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

func Test_limitSymbolResults(t *testing.T) {
	t.Run("empty case => unchanged", func(t *testing.T) {
		var res []*fileMatchResolver
		res2, limitHit := limitSymbolResults(res, 0)
		if len(res2) != 0 {
			t.Errorf("res2 = %+v, want empty", res2)
		}
		if limitHit {
			t.Error("limitHit is true, but the limit should not have been hit")
		}
	})

	t.Run("one file match, one symbol", func(t *testing.T) {
		res := []*fileMatchResolver{
			{
				symbols: []*searchSymbolResult{
					{
						symbol: protocol.Symbol{
							Name: fmt.Sprintf("symbol-name-%d", rand.Int()),
						},
					},
				},
			},
		}

		t.Run("limit 0 => no file matches", func(t *testing.T) {
			res2, limitHit := limitSymbolResults(res, 0)
			if len(res2) != 0 {
				t.Errorf("res2 = %+v, want empty", res2)
			}
			if !limitHit {
				t.Error("limitHit is false, but the limit should have been hit")
			}
		})

		t.Run("limit 1 => unchanged", func(t *testing.T) {
			res2, limitHit := limitSymbolResults(res, 1)
			if !reflect.DeepEqual(res2, res) {
				t.Errorf("res2 = %+v, want %+v", res2, res)
			}
			if limitHit {
				t.Error("limitHit is true")
			}
		})
	})
}

package graphqlbackend

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/gituri"
	"github.com/sourcegraph/sourcegraph/internal/symbols/protocol"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
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
	commit := &GitCommitResolver{
		repo:   &RepositoryResolver{repo: &types.Repo{ID: 1, Name: "repo"}},
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

func Test_limitingSymbolResults(t *testing.T) {
	t.Run("empty case", func(t *testing.T) {
		var res []*FileMatchResolver

		t.Run("symbol count is 0", func(t *testing.T) {
			nsym := symbolCount(res)
			if nsym != 0 {
				t.Errorf("symbolCount(res) = %d, want 0", nsym)
			}
		})

		t.Run("limiting does not change", func(t *testing.T) {
			res2 := limitSymbolResults(res, 0)
			if len(res2) != 0 {
				t.Errorf("res2 = %+v, want empty", res2)
			}
		})
	})

	t.Run("one file match, one symbol", func(t *testing.T) {
		res := []*FileMatchResolver{
			{
				symbols: []*searchSymbolResult{
					{
						symbol: protocol.Symbol{
							Name: "symbol-name-1",
						},
					},
				},
			},
		}

		t.Run("symbol count is 1", func(t *testing.T) {
			nsym := symbolCount(res)
			if nsym != 1 {
				t.Errorf("symbolCount(res) = %d, want 1", nsym)
			}
		})

		t.Run("limit 0 => no file matches", func(t *testing.T) {
			res2 := limitSymbolResults(res, 0)
			if len(res2) != 0 {
				t.Errorf("res2 = %+v, want empty", res2)
			}
		})

		t.Run("limit 1 => unchanged", func(t *testing.T) {
			res2 := limitSymbolResults(res, 1)
			if !reflect.DeepEqual(res2, res) {
				t.Errorf("res2 = %+v, want %+v", res2, res)
			}
		})
	})

	t.Run("two file matches, one symbol per file", func(t *testing.T) {
		res := []*FileMatchResolver{
			{
				symbols: []*searchSymbolResult{
					{
						symbol: protocol.Symbol{
							Name: "symbol-name-1",
						},
					},
				},
			},
			{
				symbols: []*searchSymbolResult{
					{
						symbol: protocol.Symbol{
							Name: "symbol-name-2",
						},
					},
				},
			},
		}

		t.Run("symbol count is 2", func(t *testing.T) {
			nsym := symbolCount(res)
			if nsym != 2 {
				t.Errorf("symbolCount(res) = %d, want 2", nsym)
			}
		})

		t.Run("limit 0 => no file matches", func(t *testing.T) {
			res2 := limitSymbolResults(res, 0)
			if len(res2) != 0 {
				t.Errorf("res2 = %+v, want empty", res2)
			}
		})

		t.Run("limit 1 => one file match", func(t *testing.T) {
			wantRes2 := res[:1]
			res2 := limitSymbolResults(res, 1)
			if !reflect.DeepEqual(res2, wantRes2) {
				t.Errorf("res2 = %+v, want %+v", res2, wantRes2)
			}
		})

		t.Run("limit 2 => unchanged", func(t *testing.T) {
			res2 := limitSymbolResults(res, 2)
			if !reflect.DeepEqual(res2, res) {
				t.Errorf("res2 = %+v, want %+v", res2, res)
			}
		})
	})

	t.Run("two file matches, multiple symbols per file", func(t *testing.T) {
		res := []*FileMatchResolver{
			{
				symbols: []*searchSymbolResult{
					{symbol: protocol.Symbol{Name: "symbol-name-1"}},
					{symbol: protocol.Symbol{Name: "symbol-name-2"}},
				},
			},
			{
				symbols: []*searchSymbolResult{
					{symbol: protocol.Symbol{Name: "symbol-name-3"}},
					{symbol: protocol.Symbol{Name: "symbol-name-4"}},
				},
			},
		}

		t.Run("symbol count is 4", func(t *testing.T) {
			nsym := symbolCount(res)
			if nsym != 4 {
				t.Errorf("symbolCount(res) = %d, want 2", nsym)
			}
		})

		testCases := []struct {
			name  string
			limit int
			want  []*FileMatchResolver
		}{
			{
				name: "limit 0 => no file matches",
				want: []*FileMatchResolver{},
			},
			{
				name:  "limit 1 => one file match with one symbol",
				limit: 1,
				want: []*FileMatchResolver{
					{
						symbols: []*searchSymbolResult{
							{symbol: protocol.Symbol{Name: "symbol-name-1"}},
						},
					},
				},
			},
			{
				name:  "limit 2 => one file match with all symbols",
				limit: 2,
				want: []*FileMatchResolver{
					{
						symbols: []*searchSymbolResult{
							{symbol: protocol.Symbol{Name: "symbol-name-1"}},
							{symbol: protocol.Symbol{Name: "symbol-name-2"}},
						},
					},
				},
			},
			{
				name:  "limit 3 => two file matches with three symbols",
				limit: 3,
				want: []*FileMatchResolver{
					{
						symbols: []*searchSymbolResult{
							{symbol: protocol.Symbol{Name: "symbol-name-1"}},
							{symbol: protocol.Symbol{Name: "symbol-name-2"}},
						},
					},
					{
						symbols: []*searchSymbolResult{
							{symbol: protocol.Symbol{Name: "symbol-name-3"}},
						},
					},
				},
			},
			{
				name:  "limit 4 => two file matches with all symbols",
				limit: 4,
				want: []*FileMatchResolver{
					{
						symbols: []*searchSymbolResult{
							{symbol: protocol.Symbol{Name: "symbol-name-1"}},
							{symbol: protocol.Symbol{Name: "symbol-name-2"}},
						},
					},
					{
						symbols: []*searchSymbolResult{
							{symbol: protocol.Symbol{Name: "symbol-name-3"}},
							{symbol: protocol.Symbol{Name: "symbol-name-4"}},
						},
					},
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				res2 := limitSymbolResults(res, tc.limit)
				if len(res2) != len(tc.want) {
					t.Errorf("len(res2)=%d, len(want)=%d", len(res2), len(tc.want))
				}

				if !reflect.DeepEqual(res2, tc.want) {
					t.Error(cmp.Diff(res2, tc.want))
				}
			})
		}
	})
}

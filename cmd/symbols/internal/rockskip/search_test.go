package rockskip

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestIsFileExtensionMatch(t *testing.T) {
	tests := []struct {
		regex string
		want  []string
	}{
		{
			regex: "\\.(go)",
			want:  nil,
		},
		{
			regex: "(go)$",
			want:  nil,
		},
		{
			regex: "\\.(go)$",
			want:  []string{"go"},
		},
		{
			regex: "\\.(ts|tsx)$",
			want:  []string{"ts", "tsx"},
		},
	}
	for _, test := range tests {
		got := isFileExtensionMatch(test.regex)
		if diff := cmp.Diff(got, test.want); diff != "" {
			t.Fatalf("isFileExtensionMatch(%q) returned %v, want %v, diff: %s", test.regex, got, test.want, diff)
		}
	}
}

func TestIsLiteralPrefix(t *testing.T) {
	tests := []struct {
		expr   string
		prefix *string
	}{
		{``, nil},
		{`^`, pointers.Ptr(``)},
		{`^foo`, pointers.Ptr(`foo`)},
		{`^foo/bar\.go`, pointers.Ptr(`foo/bar.go`)},
		{`foo/bar\.go`, nil},
	}

	for _, test := range tests {
		prefix, isPrefix, err := isLiteralPrefix(test.expr)
		if err != nil {
			t.Fatal(err)
		}

		if test.prefix == nil {
			if isPrefix {
				t.Fatalf("expected isLiteralPrefix(%q) to return false", test.expr)
			}
			continue
		}

		if prefix != *test.prefix {
			t.Errorf("isLiteralPrefix(%q) = %v, want %v", test.expr, prefix, *test.prefix)
		}
	}
}

func TestSearch(t *testing.T) {
	repo, repoDir := gitserver.MakeGitRepositoryAndReturnDir(t)
	// Needed in CI
	gitRun(t, repoDir, "config", "user.email", "test@sourcegraph.com")

	git, err := newSubprocessGit(t, repoDir)
	require.NoError(t, err)
	defer git.Close()

	db, s := mockService(t, git)
	defer db.Close()

	// We don't use 'state' in this test, it's just needed for these method calls.
	state := map[string][]string{}
	gitAdd(t, repoDir, state, "a.txt", "sym1a\n")
	gitAdd(t, repoDir, state, "b.txt", "sym1b\n")
	gitRm(t, repoDir, state, "a.txt")

	out, err := gitserver.CreateGitCommand(repoDir, "git", "rev-parse", "HEAD").CombinedOutput()
	require.NoError(t, err, string(out))
	commit1 := string(bytes.TrimSpace(out))

	gitAdd(t, repoDir, state, "c.txt", "sym1c\nsym2c")
	gitAdd(t, repoDir, state, "a.java", "sym1a\nsym2a")
	gitAdd(t, repoDir, state, "b.java", "sym1b\nsym2b")
	gitRm(t, repoDir, state, "a.java")
	gitAdd(t, repoDir, state, "a.txt", "sym2a\nsym3a")

	gitAdd(t, repoDir, state, "a.go", "sym1a\nsym2a")

	out, err = gitserver.CreateGitCommand(repoDir, "git", "rev-parse", "HEAD").CombinedOutput()
	require.NoError(t, err, string(out))
	commit2 := string(bytes.TrimSpace(out))

	tests := []struct {
		name    string
		params  search.SymbolsParameters
		want    result.Symbols
		wantErr bool
	}{
		{
			name: "current commit",
			params: search.SymbolsParameters{
				Repo:     repo,
				CommitID: api.CommitID(commit2),
				Query:    "sym",
			},
			want: result.Symbols{
				result.Symbol{Path: "b.txt", Name: "sym1b"},
				result.Symbol{Path: "c.txt", Name: "sym1c"},
				result.Symbol{Path: "c.txt", Name: "sym2c", Line: 1},
				result.Symbol{Path: "b.java", Name: "sym1b"},
				result.Symbol{Path: "b.java", Name: "sym2b", Line: 1},
				result.Symbol{Path: "a.txt", Name: "sym2a"},
				result.Symbol{Path: "a.txt", Name: "sym3a", Line: 1},
				result.Symbol{Path: "a.go", Name: "sym1a"},
				result.Symbol{Path: "a.go", Name: "sym2a", Line: 1},
			},
		},
		{
			name: "current commit, no results",
			params: search.SymbolsParameters{
				Repo:     repo,
				CommitID: api.CommitID(commit2),
				Query:    "nonexistent",
			},
			want: result.Symbols{},
		},
		{
			name: "older commit",
			params: search.SymbolsParameters{
				Repo:     repo,
				CommitID: api.CommitID(commit1),
				Query:    "sym",
			},
			want: result.Symbols{
				result.Symbol{Path: "b.txt", Name: "sym1b"},
			},
		},
		{
			name: "case sensitive",
			params: search.SymbolsParameters{
				Repo:            repo,
				CommitID:        api.CommitID(commit2),
				Query:           "SYM",
				IsCaseSensitive: true,
			},
			want: result.Symbols{},
		},
		{
			name: "case insensitive",
			params: search.SymbolsParameters{
				Repo:            repo,
				CommitID:        api.CommitID(commit2),
				Query:           "SYM",
				IsCaseSensitive: false,
			},
			want: result.Symbols{
				result.Symbol{Path: "b.txt", Name: "sym1b"},
				result.Symbol{Path: "c.txt", Name: "sym1c"},
				result.Symbol{Path: "c.txt", Name: "sym2c", Line: 1},
				result.Symbol{Path: "b.java", Name: "sym1b"},
				result.Symbol{Path: "b.java", Name: "sym2b", Line: 1},
				result.Symbol{Path: "a.txt", Name: "sym2a"},
				result.Symbol{Path: "a.txt", Name: "sym3a", Line: 1},
				result.Symbol{Path: "a.go", Name: "sym1a"},
				result.Symbol{Path: "a.go", Name: "sym2a", Line: 1},
			},
		},
		{
			name: "regexp",
			params: search.SymbolsParameters{
				Repo:     repo,
				CommitID: api.CommitID(commit2),
				Query:    "sym[0-9][bc]",
				IsRegExp: true,
			},
			want: result.Symbols{
				result.Symbol{Path: "b.txt", Name: "sym1b"},
				result.Symbol{Path: "c.txt", Name: "sym1c"},
				result.Symbol{Path: "c.txt", Name: "sym2c", Line: 1},
				result.Symbol{Path: "b.java", Name: "sym1b"},
				result.Symbol{Path: "b.java", Name: "sym2b", Line: 1},
			},
		},
		{
			name: "regexp, with IsRegExp false",
			params: search.SymbolsParameters{
				Repo:     repo,
				CommitID: api.CommitID(commit2),
				Query:    "sym[0-9][bc]",
				IsRegExp: false,
			},
			want: result.Symbols{},
		},
		{
			name: "invalid regexp",
			params: search.SymbolsParameters{
				Repo:     repo,
				CommitID: api.CommitID(commit2),
				Query:    "[0-9",
				IsRegExp: true,
			},
			wantErr: true,
		},
		{
			name: "include patterns",
			params: search.SymbolsParameters{
				Repo:            repo,
				CommitID:        api.CommitID(commit2),
				Query:           "sym",
				IncludePatterns: []string{".*\\.txt"},
			},
			want: result.Symbols{
				result.Symbol{Path: "b.txt", Name: "sym1b"},
				result.Symbol{Path: "c.txt", Name: "sym1c"},
				result.Symbol{Path: "c.txt", Name: "sym2c", Line: 1},
				result.Symbol{Path: "a.txt", Name: "sym2a"},
				result.Symbol{Path: "a.txt", Name: "sym3a", Line: 1},
			},
		},
		{
			name: "exclude patterns",
			params: search.SymbolsParameters{
				Repo:           repo,
				CommitID:       api.CommitID(commit2),
				Query:          "sym",
				ExcludePattern: ".*\\.java",
			},
			want: result.Symbols{
				result.Symbol{Path: "b.txt", Name: "sym1b"},
				result.Symbol{Path: "c.txt", Name: "sym1c"},
				result.Symbol{Path: "c.txt", Name: "sym2c", Line: 1},
				result.Symbol{Path: "a.txt", Name: "sym2a"},
				result.Symbol{Path: "a.txt", Name: "sym3a", Line: 1},
				result.Symbol{Path: "a.go", Name: "sym1a"},
				result.Symbol{Path: "a.go", Name: "sym2a", Line: 1},
			},
		},
		{
			name: "include langs",
			params: search.SymbolsParameters{
				Repo:         repo,
				CommitID:     api.CommitID(commit2),
				Query:        "sym",
				IncludeLangs: []string{"java"},
			},
			want: result.Symbols{
				result.Symbol{Path: "b.java", Name: "sym1b"},
				result.Symbol{Path: "b.java", Name: "sym2b", Line: 1},
			},
		},
		{
			name: "exclude langs",
			params: search.SymbolsParameters{
				Repo:         repo,
				CommitID:     api.CommitID(commit2),
				Query:        "sym",
				ExcludeLangs: []string{"java", "txt"},
			},
			want: result.Symbols{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			got, err := s.Search(ctx, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("Unexpected number of results: got %d, want %d", len(got), len(tt.want))
			}

			for _, w := range tt.want {
				found := false
				for _, g := range got {
					if reflect.DeepEqual(g, w) {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("Expected result missing: %v", w)
				}
			}
		})
	}
}

func TestSearchResultLimiting(t *testing.T) {
	repo, repoDir := gitserver.MakeGitRepositoryAndReturnDir(t)
	// Needed in CI
	gitRun(t, repoDir, "config", "user.email", "test@sourcegraph.com")

	git, err := newSubprocessGit(t, repoDir)
	require.NoError(t, err)
	defer git.Close()

	db, s := mockService(t, git)
	defer db.Close()

	// We don't use 'state' in this test, it's just needed for these method calls.
	state := map[string][]string{}
	gitAdd(t, repoDir, state, "a.txt", "sym1a\n")
	gitAdd(t, repoDir, state, "b.txt", "sym1b\n")
	gitAdd(t, repoDir, state, "c.txt", "sym1c\n")
	gitAdd(t, repoDir, state, "d.txt", "sym1d\n")
	gitRm(t, repoDir, state, "a.txt")
	gitAdd(t, repoDir, state, "a.java", "sym1a\n")
	gitAdd(t, repoDir, state, "b.java", "sym1b\n")

	out, err := gitserver.CreateGitCommand(repoDir, "git", "rev-parse", "HEAD").CombinedOutput()
	require.NoError(t, err, string(out))
	commit := string(bytes.TrimSpace(out))

	tests := []struct {
		name      string
		params    search.SymbolsParameters
		wantCount int
	}{

		{
			name: "limit results",
			params: search.SymbolsParameters{
				Repo:     repo,
				CommitID: api.CommitID(commit),
				Query:    "sym",
				First:    2,
			},
			wantCount: 2,
		},
		{
			name: "no limit",
			params: search.SymbolsParameters{
				Repo:     repo,
				CommitID: api.CommitID(commit),
				Query:    "sym",
			},
			wantCount: 5,
		},
		{
			name: "limit results with patterns",
			params: search.SymbolsParameters{
				Repo:            repo,
				CommitID:        api.CommitID(commit),
				Query:           "sym",
				First:           1,
				IncludePatterns: []string{".*\\.java"},
			},
			wantCount: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			symbols, err := s.Search(ctx, test.params)
			require.NoError(t, err)
			require.Equal(t, test.wantCount, len(symbols))
		})
	}
}

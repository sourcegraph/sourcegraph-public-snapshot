package search

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

// initGitRepository initializes a new Git repository and runs cmds in a new
// temporary directory (returned as dir).
func initGitRepository(t testing.TB, cmds ...string) string {
	t.Helper()
	dir := t.TempDir()
	cmds = append([]string{"git init"}, cmds...)
	for _, cmd := range cmds {
		out, err := gitCommand(dir, "bash", "-c", cmd).CombinedOutput()
		if err != nil {
			t.Fatalf("Command %q failed. Output was:\n\n%s", cmd, out)
		}
	}
	return dir
}

func gitCommand(dir, name string, args ...string) *exec.Cmd {
	c := exec.Command(name, args...)
	c.Dir = dir
	c.Env = append(os.Environ(), "GIT_CONFIG="+path.Join(dir, ".git", "config"))
	return c
}

func TestSearch(t *testing.T) {
	cmds := []string{
		"echo lorem ipsum dolor sit amet > file1",
		"git add -A",
		"GIT_COMMITTER_NAME=camden1 " +
			"GIT_COMMITTER_EMAIL=camden1@ccheek.com " +
			"GIT_COMMITTER_DATE=2006-01-02T15:04:05Z " +
			"GIT_AUTHOR_NAME=camden1 " +
			"GIT_AUTHOR_EMAIL=camden1@ccheek.com " +
			"GIT_AUTHOR_DATE=2006-01-02T15:04:05Z " +
			"git commit -m commit1 ",
		"echo consectetur adipiscing elit > file2",
		"git add -A",
		"GIT_COMMITTER_NAME=camden2 " +
			"GIT_COMMITTER_EMAIL=camden2@ccheek.com " +
			"GIT_COMMITTER_DATE=2006-01-02T15:04:05Z " +
			"GIT_AUTHOR_NAME=camden2 " +
			"GIT_AUTHOR_EMAIL=camden2@ccheek.com " +
			"GIT_AUTHOR_DATE=2006-01-02T15:04:05Z " +
			"git commit -m commit2",
	}
	dir := initGitRepository(t, cmds...)

	t.Run("match one", func(t *testing.T) {
		query := &protocol.MessageMatches{Expr: "commit2"}
		tree, err := ToMatchTree(query)
		require.NoError(t, err)
		searcher := &CommitSearcher{
			RepoDir: dir,
			Query:   tree,
		}
		var matches []*protocol.CommitMatch
		err = searcher.Search(context.Background(), func(match *protocol.CommitMatch) {
			matches = append(matches, match)
		})
		require.NoError(t, err)
		require.Len(t, matches, 1)
	})

	t.Run("match both, in order", func(t *testing.T) {
		query := &protocol.MessageMatches{Expr: "c"}
		tree, err := ToMatchTree(query)
		require.NoError(t, err)
		searcher := &CommitSearcher{
			RepoDir: dir,
			Query:   tree,
		}
		var matches []*protocol.CommitMatch
		err = searcher.Search(context.Background(), func(match *protocol.CommitMatch) {
			matches = append(matches, match)
		})
		require.NoError(t, err)
		require.Len(t, matches, 2)
		require.Equal(t, matches[0].Author.Name, "camden2")
		require.Equal(t, matches[1].Author.Name, "camden1")
	})

	t.Run("match diff content", func(t *testing.T) {
		query := &protocol.DiffMatches{Expr: "ipsum"}
		tree, err := ToMatchTree(query)
		require.NoError(t, err)
		searcher := &CommitSearcher{
			RepoDir: dir,
			Query:   tree,
		}
		var matches []*protocol.CommitMatch
		err = searcher.Search(context.Background(), func(match *protocol.CommitMatch) {
			matches = append(matches, match)
		})
		require.NoError(t, err)
		require.Len(t, matches, 1)
		require.Equal(t, matches[0].Author.Name, "camden1")
	})

	t.Run("author matches", func(t *testing.T) {
		query := &protocol.AuthorMatches{Expr: "2"}
		tree, err := ToMatchTree(query)
		require.NoError(t, err)
		searcher := &CommitSearcher{
			RepoDir: dir,
			Query:   tree,
		}
		var matches []*protocol.CommitMatch
		err = searcher.Search(context.Background(), func(match *protocol.CommitMatch) {
			matches = append(matches, match)
		})
		require.NoError(t, err)
		require.Len(t, matches, 1)
		require.Equal(t, matches[0].Author.Name, "camden2")
	})

	t.Run("file matches", func(t *testing.T) {
		query := &protocol.DiffModifiesFile{Expr: "file1"}
		tree, err := ToMatchTree(query)
		require.NoError(t, err)
		searcher := &CommitSearcher{
			RepoDir: dir,
			Query:   tree,
		}
		var matches []*protocol.CommitMatch
		err = searcher.Search(context.Background(), func(match *protocol.CommitMatch) {
			matches = append(matches, match)
		})
		require.NoError(t, err)
		require.Len(t, matches, 1)
		require.Equal(t, matches[0].Author.Name, "camden1")
	})

	t.Run("and match", func(t *testing.T) {
		query := &protocol.Operator{
			Kind: protocol.And,
			Operands: []protocol.Node{
				&protocol.DiffMatches{Expr: "lorem"},
				&protocol.DiffMatches{Expr: "ipsum"},
			},
		}
		tree, err := ToMatchTree(query)
		require.NoError(t, err)
		searcher := &CommitSearcher{
			RepoDir:     dir,
			Query:       tree,
			IncludeDiff: true,
		}
		var matches []*protocol.CommitMatch
		err = searcher.Search(context.Background(), func(match *protocol.CommitMatch) {
			matches = append(matches, match)
		})
		require.NoError(t, err)
		require.Len(t, matches, 1)
		require.Equal(t, matches[0].Author.Name, "camden1")
		require.Len(t, strings.Split(matches[0].Diff.Content, "\n"), 4)
	})
}

func TestCommitScanner(t *testing.T) {
	cases := []struct {
		input    []byte
		expected []*RawCommit
	}{{
		input: []byte(
			"2061ba96d63cba38f20a76f039cf29ef68736b8a\x00\x00HEAD\x00Camden Cheek\x00camden@sourcegraph.com\x001632251505\x00Camden Cheek\x00camden@sourcegraph.com\x001632251505\x00fix import\n\x005230097b75dcbb2c214618dd171da4053aff18a6\x00\x00" +
				"5230097b75dcbb2c214618dd171da4053aff18a6\x00\x00HEAD\x00Camden Cheek\x00camden@sourcegraph.com\x001632248499\x00Camden Cheek\x00camden@sourcegraph.com\x001632248499\x00only set matches if they exist\n\x00\x00",
		),
		expected: []*RawCommit{{
			Hash:           []byte("2061ba96d63cba38f20a76f039cf29ef68736b8a"),
			RefNames:       []byte(""),
			SourceRefs:     []byte("HEAD"),
			AuthorName:     []byte("Camden Cheek"),
			AuthorEmail:    []byte("camden@sourcegraph.com"),
			AuthorDate:     []byte("1632251505"),
			CommitterName:  []byte("Camden Cheek"),
			CommitterEmail: []byte("camden@sourcegraph.com"),
			CommitterDate:  []byte("1632251505"),
			Message:        []byte("fix import"),
			ParentHashes:   []byte("5230097b75dcbb2c214618dd171da4053aff18a6"),
		}, {
			Hash:           []byte("5230097b75dcbb2c214618dd171da4053aff18a6"),
			RefNames:       []byte(""),
			SourceRefs:     []byte("HEAD"),
			AuthorName:     []byte("Camden Cheek"),
			AuthorEmail:    []byte("camden@sourcegraph.com"),
			AuthorDate:     []byte("1632248499"),
			CommitterName:  []byte("Camden Cheek"),
			CommitterEmail: []byte("camden@sourcegraph.com"),
			CommitterDate:  []byte("1632248499"),
			Message:        []byte("only set matches if they exist"),
			ParentHashes:   []byte(""),
		}},
	}}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			scanner := NewCommitScanner(bytes.NewReader(tc.input))
			var output []*RawCommit
			for scanner.Scan() {
				output = append(output, scanner.NextRawCommit())
			}
			require.NoError(t, scanner.Err())
			require.Equal(t, tc.expected, output)
		})
	}
}

func TestHighlights(t *testing.T) {
	rawDiff := `diff --git internal/compute/match.go internal/compute/match.go
new file mode 100644
index 0000000000..fcc91bf673
--- /dev/null
+++ internal/compute/match.go
@@ -0,0 +1,97 @@
+package compute
+
+import (
+       "fmt"
+       "regexp"
+
+       "github.com/sourcegraph/sourcegraph/internal/search/result"
+)
+
+func ofFileMatches(fm *result.FileMatch, r *regexp.Regexp) *Result {
+       matches := make([]Match, 0, len(fm.LineMatches))
+       for _, l := range fm.LineMatches {
+               regexpMatches := r.FindAllStringSubmatchIndex(l.Preview, -1)
+               matches = append(matches, ofRegexpMatches(regexpMatches, l.Preview, int(l.LineNumber)))
+       }
+       return &Result{Matches: matches, Path: fm.Path}
+}
diff --git internal/compute/match_test.go internal/compute/match_test.go
new file mode 100644
index 0000000000..7e54670557
--- /dev/null
+++ internal/compute/match_test.go
@@ -0,0 +1,112 @@
+package compute
+
+import (
+       "encoding/json"
+       "regexp"
+       "testing"
+
+       "github.com/hexops/autogold"
+       "github.com/sourcegraph/sourcegraph/internal/search/result"
+)
+
+func TestOfLineMatches(t *testing.T) {
+       test := func(input string) string {
+               r, _ := regexp.Compile(input)
+               result := ofFileMatches(data, r)
+               v, _ := json.MarshalIndent(result, "", "  ")
+               return string(v)
+       }
+}`

	parsedDiff, err := diff.NewMultiFileDiffReader(strings.NewReader(rawDiff)).ReadAllFiles()
	require.NoError(t, err)

	lc := &LazyCommit{
		RawCommit: &RawCommit{
			AuthorName: []byte("Camden Cheek"),
		},
		diff: parsedDiff,
	}

	mt, err := ToMatchTree(&protocol.Operator{
		Kind: protocol.And,
		Operands: []protocol.Node{
			&protocol.AuthorMatches{Expr: "Camden"},
			&protocol.DiffModifiesFile{Expr: "test"},
			&protocol.Operator{
				Kind: protocol.And,
				Operands: []protocol.Node{
					&protocol.DiffMatches{Expr: "result"},
					&protocol.DiffMatches{Expr: "test"},
				},
			},
		},
	})
	require.NoError(t, err)

	matches, highlights, err := mt.Match(lc)
	require.NoError(t, err)
	require.True(t, matches)

	formatted, ranges := FormatDiff(parsedDiff, highlights.Diff)
	expectedFormatted := `/dev/null internal/compute/match.go
@@ -0,0 +6,6 @@ 
+
+       "github.com/sourcegraph/sourcegraph/internal/search/result"
+)
+
+func ofFileMatches(fm *result.FileMatch, r *regexp.Regexp) *Result {
+       matches := make([]Match, 0, len(fm.LineMatches))
/dev/null internal/compute/match_test.go
@@ -0,0 +5,7 @@ ... +4
+       "regexp"
+       "testing"
+
+       "github.com/hexops/autogold"
+       "github.com/sourcegraph/sourcegraph/internal/search/result"
+)
+
`

	require.Equal(t, expectedFormatted, formatted)

	expectedRanges := result.Ranges{{
		Start: result.Location{Offset: 115, Line: 3, Column: 60},
		End:   result.Location{Offset: 121, Line: 3, Column: 66},
	}, {
		Start: result.Location{Offset: 152, Line: 6, Column: 24},
		End:   result.Location{Offset: 158, Line: 6, Column: 30},
	}, {
		Start: result.Location{Offset: 288, Line: 8, Column: 33},
		End:   result.Location{Offset: 292, Line: 8, Column: 37},
	}, {
		Start: result.Location{Offset: 345, Line: 11, Column: 9},
		End:   result.Location{Offset: 349, Line: 11, Column: 13},
	}, {
		Start: result.Location{Offset: 453, Line: 14, Column: 60},
		End:   result.Location{Offset: 459, Line: 14, Column: 66},
	}}

	require.Equal(t, expectedRanges, ranges)

}

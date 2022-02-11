package search

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strings"
	"testing"
	"testing/quick"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
		"echo consectetur adipiscing elit again > file3",
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
		require.Empty(t, matches[0].ModifiedFiles)
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

	t.Run("and with no operands matches all", func(t *testing.T) {
		query := protocol.NewAnd()
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
		query := protocol.NewAnd(
			&protocol.DiffMatches{Expr: "lorem"},
			&protocol.DiffMatches{Expr: "ipsum"},
		)
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

	t.Run("match both, in order with modified files", func(t *testing.T) {
		query := &protocol.MessageMatches{Expr: "c"}
		tree, err := ToMatchTree(query)
		require.NoError(t, err)
		searcher := &CommitSearcher{
			RepoDir:              dir,
			Query:                tree,
			IncludeModifiedFiles: true,
		}
		var matches []*protocol.CommitMatch
		err = searcher.Search(context.Background(), func(match *protocol.CommitMatch) {
			matches = append(matches, match)
		})
		require.NoError(t, err)
		require.Len(t, matches, 2)
		require.Equal(t, matches[0].Author.Name, "camden2")
		require.Equal(t, matches[1].Author.Name, "camden1")
		require.Equal(t, []string{"file2", "file3"}, matches[0].ModifiedFiles)
		require.Equal(t, []string{"file1"}, matches[1].ModifiedFiles)
	})
}

func TestCommitScanner(t *testing.T) {
	cases := []struct {
		input    []byte
		expected []*RawCommit
	}{
		{
			input: []byte(
				"\x1E2061ba96d63cba38f20a76f039cf29ef68736b8a\x00\x00HEAD\x00Camden Cheek\x00camden@sourcegraph.com\x001632251505\x00Camden Cheek\x00camden@sourcegraph.com\x001632251505\x00fix import\n\x005230097b75dcbb2c214618dd171da4053aff18a6\x00\x00" +
					"\x1E5230097b75dcbb2c214618dd171da4053aff18a6\x00\x00HEAD\x00Camden Cheek\x00camden@sourcegraph.com\x001632248499\x00Camden Cheek\x00camden@sourcegraph.com\x001632248499\x00only set matches if they exist\n\x00\x00",
			),
			expected: []*RawCommit{
				{
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
					ModifiedFiles:  [][]byte{{}, {}},
				},
				{
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
					ModifiedFiles:  [][]byte{{}},
				},
			},
		},
		{
			input: []byte(
				"\x1E2061ba96d63cba38f20a76f039cf29ef68736b8a\x00\x00HEAD\x00Camden Cheek\x00camden@sourcegraph.com\x001632251505\x00Camden Cheek\x00camden@sourcegraph.com\x001632251505\x00fix import\n\x005230097b75dcbb2c214618dd171da4053aff18a6\x00\x00file1" +
					"\x1E5230097b75dcbb2c214618dd171da4053aff18a6\x00\x00HEAD\x00Camden Cheek\x00camden@sourcegraph.com\x001632248499\x00Camden Cheek\x00camden@sourcegraph.com\x001632248499\x00only set matches if they exist\n\x00\x00file1\x00file2",
			),
			expected: []*RawCommit{
				{
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
					ModifiedFiles: [][]byte{
						{},
						[]byte("file1"),
					},
				},
				{
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
					ModifiedFiles: [][]byte{
						[]byte("file1"),
						[]byte("file2"),
					},
				},
			},
		},
	}

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

	mt, err := ToMatchTree(protocol.NewAnd(
		&protocol.AuthorMatches{Expr: "Camden"},
		&protocol.DiffModifiesFile{Expr: "match"},
		protocol.NewOr(
			&protocol.DiffMatches{Expr: "result"},
			&protocol.DiffMatches{Expr: "test"},
		),
	))
	require.NoError(t, err)

	mergedResult, highlights, err := mt.Match(lc)
	require.NoError(t, err)
	require.True(t, mergedResult.Satisfies())

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
		Start: result.Location{Offset: 27, Line: 0, Column: 27},
		End:   result.Location{Offset: 32, Line: 0, Column: 32},
	}, {
		Start: result.Location{Offset: 115, Line: 3, Column: 60},
		End:   result.Location{Offset: 121, Line: 3, Column: 66},
	}, {
		Start: result.Location{Offset: 152, Line: 6, Column: 24},
		End:   result.Location{Offset: 158, Line: 6, Column: 30},
	}, {
		Start: result.Location{Offset: 282, Line: 8, Column: 27},
		End:   result.Location{Offset: 287, Line: 8, Column: 32},
	}, {
		Start: result.Location{Offset: 345, Line: 11, Column: 9},
		End:   result.Location{Offset: 349, Line: 11, Column: 13},
	}, {
		Start: result.Location{Offset: 453, Line: 14, Column: 60},
		End:   result.Location{Offset: 459, Line: 14, Column: 66},
	}}

	require.Equal(t, expectedRanges, ranges)

}

func TestFuzzQueryCNF(t *testing.T) {
	matchTreeMatches := func(mt MatchTree, a authorNameGenerator) bool {
		lc := &LazyCommit{
			RawCommit: &RawCommit{
				AuthorName: []byte(a),
			},
		}
		mergedResult, _, err := mt.Match(lc)
		require.NoError(t, err)
		return mergedResult.Satisfies()
	}

	rawQueryMatches := func(q queryGenerator, a authorNameGenerator) bool {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Recovered from panic in rawQueryMatches:\n  Query: %s\n  Author: %s\n", q.RawQuery.String(), a)
			}
		}()
		mt, err := ToMatchTree(q.RawQuery)
		require.NoError(t, err)
		return matchTreeMatches(mt, a)
	}

	reducedQueryMatches := func(q queryGenerator, a authorNameGenerator) bool {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Recovered from panic in reducedQueryMatches:\n  Query: %s\n  Author: %s\n", q.RawQuery.String(), a)
			}
		}()
		mt, err := ToMatchTree(q.ConstructedQuery())
		require.NoError(t, err)
		return matchTreeMatches(mt, a)
	}

	err := quick.CheckEqual(rawQueryMatches, reducedQueryMatches, nil)
	var e *quick.CheckEqualError
	if err != nil && errors.As(err, &e) {
		t.Fatalf("Different outputs for same inputs\n  RawQuery: %s\n  ReducedQuery: %s\n  AuthorName: %s\n",
			e.In[0].(queryGenerator).RawQuery.String(),
			e.In[0].(queryGenerator).ConstructedQuery().String(),
			string(e.In[1].([]uint8)),
		)
	} else if err != nil {
		t.Fatal(err)
	}
}

// queryGenerator is a type that satisfies the tesing/quick Generator interface,
// generating random, unreduced queries in its RawQuery field. Additionally,
// it exposes a ConstructedQuery() convienence method that allows the caller to get the
// query as if it had been created with the protocol.New* functions.
type queryGenerator struct {
	RawQuery protocol.Node
}

func (queryGenerator) Generate(rand *rand.Rand, size int) reflect.Value {
	// Set max depth to avoid massive trees
	if size > 10 {
		size = 10
	}
	return reflect.ValueOf(queryGenerator{generateQuery(rand, size)})
}

// ConstructedQuery returns the query as if constructted with the protocol.New* functions
func (q queryGenerator) ConstructedQuery() protocol.Node {
	return protocol.Reduce(constructedQuery(q.RawQuery))
}

// constructedQuery takes any query and recursively reduces it with the
// protocol.New* functions. This is not meant to be used outside of fuzz testing
// because any caller should be using the protocol.New* functions directly, which
// reduce the query on construction.
func constructedQuery(q protocol.Node) protocol.Node {
	switch v := q.(type) {
	case *protocol.Operator:
		newOperands := make([]protocol.Node, 0, len(v.Operands))
		for _, operand := range v.Operands {
			newOperands = append(newOperands, constructedQuery(operand))
		}
		switch v.Kind {
		case protocol.And:
			return protocol.NewAnd(newOperands...)
		case protocol.Or:
			return protocol.NewOr(newOperands...)
		case protocol.Not:
			return protocol.NewNot(newOperands[0])
		default:
			panic("unreachable")
		}
	default:
		return v
	}
}

const randomChars = `abcdefghijkl`

// generateAtom generates a random AuthorMatches atom.
// The AuthorMatches node will match a single, random character from `randomChars`.
// 50% of the generated nodes will also be negated. We negate in the atom step
// rather than in the generateQuery step because we only want to generate negated
// nodes if they are wrapping leaf nodes. Negating non-leaf nodes works correctly,
// but can lead to multiple-exponential behavior.
func generateAtom(rand *rand.Rand) protocol.Node {
	a := &protocol.AuthorMatches{
		Expr: string(randomChars[rand.Int()%len(randomChars)]),
	}
	if rand.Int()%2 == 0 {
		return a
	}
	return &protocol.Operator{Kind: protocol.Not, Operands: []protocol.Node{a}}
}

// generateQuery generates a random query with configurable depth. Atom,
// And, and Or nodes will occur with a 1:1:1 ratio on average.
func generateQuery(rand *rand.Rand, depth int) protocol.Node {
	if depth == 0 {
		return generateAtom(rand)
	}

	switch rand.Int() % 3 {
	case 0:
		var operands []protocol.Node
		for i := 0; i < rand.Int()%4; i++ {
			operands = append(operands, generateQuery(rand, depth-1))
		}
		return &protocol.Operator{Kind: protocol.And, Operands: operands}
	case 1:
		var operands []protocol.Node
		for i := 0; i < rand.Int()%4; i++ {
			operands = append(operands, generateQuery(rand, depth-1))
		}
		return &protocol.Operator{Kind: protocol.Or, Operands: operands}
	case 2:
		return generateAtom(rand)
	default:
		panic("unreachable")
	}
}

// authorNameGenerator is a type that implements the testing/quick Generator interface
// so it can be randomly generated using the same characters that the AuthorMatches
// nodes are generated with using generateAtom.
type authorNameGenerator []byte

func (authorNameGenerator) Generate(rand *rand.Rand, size int) reflect.Value {
	if size > 10 {
		size = 10
	}
	buf := make([]byte, size)
	for i := 0; i < len(buf); i++ {
		buf[i] = randomChars[rand.Int()%len(randomChars)]
	}
	return reflect.ValueOf(buf)
}

func Test_revsToGitArgs(t *testing.T) {
	cases := []struct {
		name     string
		revSpecs []protocol.RevisionSpecifier
		expected []string
	}{{
		name: "explicit HEAD",
		revSpecs: []protocol.RevisionSpecifier{{
			RevSpec: "HEAD",
		}},
		expected: []string{"HEAD"},
	}, {
		name:     "implicit HEAD",
		revSpecs: []protocol.RevisionSpecifier{{}},
		expected: []string{"HEAD"},
	}, {
		name: "glob",
		revSpecs: []protocol.RevisionSpecifier{{
			RefGlob: "refs/heads/*",
		}},
		expected: []string{"--glob=refs/heads/*"},
	}, {
		name: "glob with excluded",
		revSpecs: []protocol.RevisionSpecifier{{
			RefGlob: "refs/heads/*",
		}, {
			ExcludeRefGlob: "refs/heads/cc/*",
		}},
		expected: []string{
			"--glob=refs/heads/*",
			"--exclude=refs/heads/cc/*",
		},
	}}

	for _, tc := range cases {
		got := revsToGitArgs(tc.revSpecs)
		require.Equal(t, tc.expected, got)
	}
}

package search

// "context"
// "fmt"
// "regexp"
// "testing"

// "github.com/stretchr/testify/require"

// func TestFormatDiff(t *testing.T) {
// 	pred := And{
// 		&AuthorMatches{regexp.MustCompile(`camden`)},
// 		&DiffMatches{regexp.MustCompile(`dec\.ReadAll`)},
// 	}

// 	err := IterCommitMatches(context.Background(), "/Users/ccheek/src/sourcegraph/sourcegraph", "HEAD", pred, func(match CommitMatch) bool {
// 		diff, _ := match.Commit.Diff()
// 		formatted, ranges := FormatDiffWithHighlights(diff, match.Highlights.Diff)
// 		println(match.Commit.Id().String())
// 		print(formatted)
// 		fmt.Printf("%#v\n", ranges)
// 		return true
// 	})

// 	require.NoError(t, err)
// }

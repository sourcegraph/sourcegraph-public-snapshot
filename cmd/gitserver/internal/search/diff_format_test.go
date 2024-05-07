package search

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

func TestDiffFormat(t *testing.T) {
	t.Run("last line matches", func(t *testing.T) {
		rawDiff := `diff --git a/.mailmap b/.mailmap
index dbace57d5f..53357b4971 100644
--- file with spaces
+++ new file with spaces
@@ -59,3 +59,4 @@ Unknown <u@gogs.io> 无闻 <u@gogs.io>
 Renovate Bot <bot@renovateapp.com> renovate[bot] <renovate[bot]@users.noreply.github.com>
 Matt King <kingy895@gmail.com> Matthew King <kingy895@gmail.com>
+Camden Cheek <camden@sourcegraph.com> Camden Cheek <camden@ccheek.com>
`
		parsedDiff, err := diff.NewMultiFileDiffReader(strings.NewReader(rawDiff)).ReadAllFiles()
		require.NoError(t, err)

		highlights := map[int]MatchedFileDiff{
			0: {MatchedHunks: map[int]MatchedHunk{
				0: {MatchedLines: map[int]result.Ranges{
					2: {{
						Start: result.Location{Offset: 0, Line: 0, Column: 0},
						End:   result.Location{Offset: 6, Line: 0, Column: 6},
					}},
				}},
			}},
		}

		formatted, ranges := FormatDiff(parsedDiff, highlights)
		expectedFormatted := `file\ with\ spaces new\ file\ with\ spaces
@@ -60,1 +60,2 @@ Unknown <u@gogs.io> 无闻 <u@gogs.io>
 Matt King <kingy895@gmail.com> Matthew King <kingy895@gmail.com>
+Camden Cheek <camden@sourcegraph.com> Camden Cheek <camden@ccheek.com>
`
		require.Equal(t, expectedFormatted, formatted)

		expectedRanges := result.Ranges{{
			Start: result.Location{Offset: 167, Line: 3, Column: 1},
			End:   result.Location{Offset: 173, Line: 3, Column: 7},
		}}
		require.Equal(t, expectedRanges, ranges)
	})

	t.Run("invalid utf8", func(t *testing.T) {
		rawDiff := "diff --git a/.mailmap b/.mailmap\n" +
			"index dbace57d5f..53357b4971 100644\n" +
			"--- file with spaces\n" +
			"+++ new file with spaces and invalid\xC0 utf8\n" +
			"@@ -59,3 +59,4 @@ Unknown <u@gogs.io> 无闻 <u@gogs.io>\n" +
			" Renovate Bot <bot@renovateapp.com> renovate[bot] <renovate[bot]@users.noreply.github.com>\n" +
			" Matt King <kingy895@gmail.com> Matthew King <kingy895@gmail.com>\n" +
			// \xC0 is an invalid UTF8 byte
			"+Camden Cheek <invalid@utf8.\xC0m> Camden Cheek <camden@ccheek.com>\n"

		parsedDiff, err := diff.NewMultiFileDiffReader(strings.NewReader(rawDiff)).ReadAllFiles()
		require.NoError(t, err)

		highlights := map[int]MatchedFileDiff{
			0: {MatchedHunks: map[int]MatchedHunk{
				0: {MatchedLines: map[int]result.Ranges{
					2: {{
						Start: result.Location{Offset: 0, Line: 0, Column: 0},
						End:   result.Location{Offset: 6, Line: 0, Column: 6},
					}},
				}},
			}},
		}

		formatted, _ := FormatDiff(parsedDiff, highlights)
		expectedFormatted := "file\\ with\\ spaces new\\ file\\ with\\ spaces\\ and\\ invalid�\\ utf8\n" +
			"@@ -60,1 +60,2 @@ Unknown <u@gogs.io> 无闻 <u@gogs.io>\n" +
			" Matt King <kingy895@gmail.com> Matthew King <kingy895@gmail.com>\n" +
			"+Camden Cheek <invalid@utf8.�m> Camden Cheek <camden@ccheek.com>\n"
		require.Equal(t, expectedFormatted, formatted)
		require.True(t, utf8.ValidString(formatted))
	})
}

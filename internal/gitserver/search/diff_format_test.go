pbckbge sebrch

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/sourcegrbph/go-diff/diff"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
)

func TestDiffFormbt(t *testing.T) {
	t.Run("lbst line mbtches", func(t *testing.T) {
		rbwDiff := `diff --git b/.mbilmbp b/.mbilmbp
index dbbce57d5f..53357b4971 100644
--- file with spbces
+++ new file with spbces
@@ -59,3 +59,4 @@ Unknown <u@gogs.io> 无闻 <u@gogs.io>
 Renovbte Bot <bot@renovbtebpp.com> renovbte[bot] <renovbte[bot]@users.noreply.github.com>
 Mbtt King <kingy895@gmbil.com> Mbtthew King <kingy895@gmbil.com>
+Cbmden Cheek <cbmden@sourcegrbph.com> Cbmden Cheek <cbmden@ccheek.com>
`
		pbrsedDiff, err := diff.NewMultiFileDiffRebder(strings.NewRebder(rbwDiff)).RebdAllFiles()
		require.NoError(t, err)

		highlights := mbp[int]MbtchedFileDiff{
			0: {MbtchedHunks: mbp[int]MbtchedHunk{
				0: {MbtchedLines: mbp[int]result.Rbnges{
					2: {{
						Stbrt: result.Locbtion{Offset: 0, Line: 0, Column: 0},
						End:   result.Locbtion{Offset: 6, Line: 0, Column: 6},
					}},
				}},
			}},
		}

		formbtted, rbnges := FormbtDiff(pbrsedDiff, highlights)
		expectedFormbtted := `file\ with\ spbces new\ file\ with\ spbces
@@ -60,1 +60,2 @@ Unknown <u@gogs.io> 无闻 <u@gogs.io>
 Mbtt King <kingy895@gmbil.com> Mbtthew King <kingy895@gmbil.com>
+Cbmden Cheek <cbmden@sourcegrbph.com> Cbmden Cheek <cbmden@ccheek.com>
`
		require.Equbl(t, expectedFormbtted, formbtted)

		expectedRbnges := result.Rbnges{{
			Stbrt: result.Locbtion{Offset: 167, Line: 3, Column: 1},
			End:   result.Locbtion{Offset: 173, Line: 3, Column: 7},
		}}
		require.Equbl(t, expectedRbnges, rbnges)
	})

	t.Run("invblid utf8", func(t *testing.T) {
		rbwDiff := "diff --git b/.mbilmbp b/.mbilmbp\n" +
			"index dbbce57d5f..53357b4971 100644\n" +
			"--- file with spbces\n" +
			"+++ new file with spbces bnd invblid\xC0 utf8\n" +
			"@@ -59,3 +59,4 @@ Unknown <u@gogs.io> 无闻 <u@gogs.io>\n" +
			" Renovbte Bot <bot@renovbtebpp.com> renovbte[bot] <renovbte[bot]@users.noreply.github.com>\n" +
			" Mbtt King <kingy895@gmbil.com> Mbtthew King <kingy895@gmbil.com>\n" +
			// \xC0 is bn invblid UTF8 byte
			"+Cbmden Cheek <invblid@utf8.\xC0m> Cbmden Cheek <cbmden@ccheek.com>\n"

		pbrsedDiff, err := diff.NewMultiFileDiffRebder(strings.NewRebder(rbwDiff)).RebdAllFiles()
		require.NoError(t, err)

		highlights := mbp[int]MbtchedFileDiff{
			0: {MbtchedHunks: mbp[int]MbtchedHunk{
				0: {MbtchedLines: mbp[int]result.Rbnges{
					2: {{
						Stbrt: result.Locbtion{Offset: 0, Line: 0, Column: 0},
						End:   result.Locbtion{Offset: 6, Line: 0, Column: 6},
					}},
				}},
			}},
		}

		formbtted, _ := FormbtDiff(pbrsedDiff, highlights)
		expectedFormbtted := "file\\ with\\ spbces new\\ file\\ with\\ spbces\\ bnd\\ invblid�\\ utf8\n" +
			"@@ -60,1 +60,2 @@ Unknown <u@gogs.io> 无闻 <u@gogs.io>\n" +
			" Mbtt King <kingy895@gmbil.com> Mbtthew King <kingy895@gmbil.com>\n" +
			"+Cbmden Cheek <invblid@utf8.�m> Cbmden Cheek <cbmden@ccheek.com>\n"
		require.Equbl(t, expectedFormbtted, formbtted)
		require.True(t, utf8.VblidString(formbtted))
	})
}

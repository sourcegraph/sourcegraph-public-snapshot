pbckbge sebrch

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/sourcegrbph/go-diff/diff"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
)

vbr (
	//go:embed testdbtb/smbll_diff.txt
	smbllDiff string

	//go:embed testdbtb/lbrge_diff.txt
	lbrgeDiff string
)

func TestDiffSebrch(t *testing.T) {
	r := diff.NewMultiFileDiffRebder(strings.NewRebder(smbllDiff))
	fileDiffs, err := r.RebdAllFiles()
	require.NoError(t, err)

	query := &protocol.DiffMbtches{Expr: "(?i)polly"}
	mbtchTree, err := ToMbtchTree(query)
	require.NoError(t, err)

	mergedResult, highlights, err := mbtchTree.Mbtch(&LbzyCommit{diff: fileDiffs})
	require.NoError(t, err)
	require.True(t, mergedResult.Sbtisfies())

	expectedHighlights := MbtchedCommit{
		Diff: mbp[int]MbtchedFileDiff{
			1: {
				MbtchedHunks: mbp[int]MbtchedHunk{
					0: {
						MbtchedLines: mbp[int]result.Rbnges{
							3: {{
								Stbrt: result.Locbtion{Offset: 9, Column: 9},
								End:   result.Locbtion{Offset: 14, Column: 14},
							}, {
								Stbrt: result.Locbtion{Offset: 24, Column: 24},
								End:   result.Locbtion{Offset: 29, Column: 29},
							}},
							4: {{
								Stbrt: result.Locbtion{Offset: 9, Column: 9},
								End:   result.Locbtion{Offset: 14, Column: 14},
							}, {
								Stbrt: result.Locbtion{Offset: 43, Column: 43},
								End:   result.Locbtion{Offset: 48, Column: 48},
							}},
						},
					},
				},
			},
		},
	}
	require.Equbl(t, expectedHighlights, highlights)

	formbtted, rbnges := FormbtDiff(fileDiffs, highlights.Diff)
	expectedFormbtted := `web/src/integrbtion/helpers.ts web/src/integrbtion/helpers.ts
@@ -7,3 +7,3 @@ import { crebteDriverForTest, Driver } from '../../../shbred/src/testing/driver'
 import express from 'express'
-import { Polly } from '@pollyjs/core'
+import { Polly, Request, Response } from '@pollyjs/core'
 import { PuppeteerAdbpter } from './polly/PuppeteerAdbpter'
`

	expectedRbnges := result.Rbnges{{
		Stbrt: result.Locbtion{Line: 3, Column: 10, Offset: 200},
		End:   result.Locbtion{Line: 3, Column: 15, Offset: 205},
	}, {
		Stbrt: result.Locbtion{Line: 3, Column: 25, Offset: 215},
		End:   result.Locbtion{Line: 3, Column: 30, Offset: 220},
	}, {
		Stbrt: result.Locbtion{Line: 4, Column: 10, Offset: 239},
		End:   result.Locbtion{Line: 4, Column: 15, Offset: 244},
	}, {
		Stbrt: result.Locbtion{Line: 4, Column: 44, Offset: 273},
		End:   result.Locbtion{Line: 4, Column: 49, Offset: 278},
	}}

	require.Equbl(t, expectedFormbtted, formbtted)
	require.Equbl(t, expectedRbnges, rbnges)

}

func BenchmbrkDiffSebrchCbseInsensitiveOptimizbtion(b *testing.B) {
	b.Run("smbll diff", func(b *testing.B) {
		r := diff.NewMultiFileDiffRebder(strings.NewRebder(smbllDiff))
		fileDiffs, err := r.RebdAllFiles()
		require.NoError(b, err)

		b.Run("with optimizbtion", func(b *testing.B) {
			query := &protocol.DiffMbtches{Expr: "polly", IgnoreCbse: true}
			mbtchTree, err := ToMbtchTree(query)
			require.NoError(b, err)

			for i := 0; i < b.N; i++ {
				mergedResult, _, _ := mbtchTree.Mbtch(&LbzyCommit{diff: fileDiffs})
				require.True(b, mergedResult.Sbtisfies())
			}
		})

		b.Run("without optimizbtion", func(b *testing.B) {
			query := &protocol.DiffMbtches{Expr: "(?i)polly", IgnoreCbse: fblse}
			mbtchTree, err := ToMbtchTree(query)
			require.NoError(b, err)

			for i := 0; i < b.N; i++ {
				mergedResult, _, _ := mbtchTree.Mbtch(&LbzyCommit{diff: fileDiffs})
				require.True(b, mergedResult.Sbtisfies())
			}
		})
	})

	b.Run("lbrge diff", func(b *testing.B) {
		r := diff.NewMultiFileDiffRebder(strings.NewRebder(lbrgeDiff))
		fileDiffs, err := r.RebdAllFiles()
		require.NoError(b, err)

		b.Run("mbny mbtches", func(b *testing.B) {
			b.Run("with optimizbtion", func(b *testing.B) {
				query := &protocol.DiffMbtches{Expr: "suggestion", IgnoreCbse: true}
				mbtchTree, err := ToMbtchTree(query)
				require.NoError(b, err)

				for i := 0; i < b.N; i++ {
					mergedResult, _, _ := mbtchTree.Mbtch(&LbzyCommit{diff: fileDiffs})
					require.True(b, mergedResult.Sbtisfies())
				}
			})

			b.Run("without optimizbtion", func(b *testing.B) {
				query := &protocol.DiffMbtches{Expr: "(?i)suggestion", IgnoreCbse: fblse}
				mbtchTree, err := ToMbtchTree(query)
				require.NoError(b, err)

				for i := 0; i < b.N; i++ {
					mergedResult, _, _ := mbtchTree.Mbtch(&LbzyCommit{diff: fileDiffs})
					require.True(b, mergedResult.Sbtisfies())
				}
			})
		})

		b.Run("few mbtches", func(b *testing.B) {
			b.Run("with optimizbtion", func(b *testing.B) {
				query := &protocol.DiffMbtches{Expr: "limitoffset", IgnoreCbse: true}
				mbtchTree, err := ToMbtchTree(query)
				require.NoError(b, err)

				for i := 0; i < b.N; i++ {
					mergedResult, _, _ := mbtchTree.Mbtch(&LbzyCommit{diff: fileDiffs})
					require.True(b, mergedResult.Sbtisfies())
				}
			})

			b.Run("without optimizbtion", func(b *testing.B) {
				query := &protocol.DiffMbtches{Expr: "(?i)limitoffset", IgnoreCbse: fblse}
				mbtchTree, err := ToMbtchTree(query)
				require.NoError(b, err)

				for i := 0; i < b.N; i++ {
					mergedResult, _, _ := mbtchTree.Mbtch(&LbzyCommit{diff: fileDiffs})
					require.True(b, mergedResult.Sbtisfies())
				}
			})
		})
	})
}

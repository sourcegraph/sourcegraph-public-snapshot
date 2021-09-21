package search

import (
	"regexp"
	"strings"
	"testing"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

func TestDiffSearch(t *testing.T) {
	rawDiff := `diff --git a/web/src/integration/gqlresponses/user_settings_bla_response_1.ts b/web/src/integration/gqlresponses/user_settings_bla_response_1.ts
new file mode 100644
index 0000000000..4f6e758628
--- /dev/null
+++ web/src/integration/gqlresponses/user_settings_bla_response_1.ts
@@ -0,0 +1,4 @@
+export const overrideSettingsResponse: OverrideSettingsResponseShape = {
+    foo: 1,
+    bar: {},
+}
diff --git a/web/src/integration/helpers.ts b/web/src/integration/helpers.ts
index 2f71392b2f..d874527291 100644
--- web/src/integration/helpers.ts
+++ web/src/integration/helpers.ts
@@ -5,7 +5,7 @@ import { createDriverForTest, Driver } from '../../../shared/src/testing/driver'
 import * as path from 'path'
 import mkdirp from 'mkdirp-promise'
 import express from 'express'
-import { Polly } from '@pollyjs/core'
+import { Polly, Request, Response } from '@pollyjs/core'
 import { PuppeteerAdapter } from './polly/PuppeteerAdapter'
 import FSPersister from '@pollyjs/persister-fs'
`
	r := diff.NewMultiFileDiffReader(strings.NewReader(rawDiff))
	fileDiffs, err := r.ReadAllFiles()
	require.NoError(t, err)

	query := &protocol.DiffMatches{protocol.Regexp{regexp.MustCompile("(?i)polly")}}
	matchTree := ToMatchTree(query)

	matched, highlights, err := matchTree.Match(&LazyCommit{diff: fileDiffs})
	require.NoError(t, err)
	require.True(t, matched)

	expectedHighlights := &protocol.CommitHighlights{
		Diff: map[int]protocol.FileDiffHighlight{
			1: protocol.FileDiffHighlight{
				HunkHighlights: map[int]protocol.HunkHighlight{
					0: protocol.HunkHighlight{
						LineHighlights: map[int]protocol.Ranges{
							3: protocol.Ranges{{
								Start: protocol.Location{Offset: 9, Column: 9},
								End:   protocol.Location{Offset: 14, Column: 14},
							}, {
								Start: protocol.Location{Offset: 24, Column: 24},
								End:   protocol.Location{Offset: 29, Column: 29},
							}},
							4: protocol.Ranges{{
								Start: protocol.Location{Offset: 9, Column: 9},
								End:   protocol.Location{Offset: 14, Column: 14},
							}, {
								Start: protocol.Location{Offset: 43, Column: 43},
								End:   protocol.Location{Offset: 48, Column: 48},
							}},
						},
					},
				},
			},
		},
	}
	require.Equal(t, expectedHighlights, highlights)

	formatted, ranges := FormatDiff(fileDiffs, highlights.Diff)
	expectedFormatted := `web/src/integration/helpers.ts web/src/integration/helpers.ts
@@ -7,3 +7,3 @@ import { createDriverForTest, Driver } from '../../../shared/src/testing/driver'
 import express from 'express'
-import { Polly } from '@pollyjs/core'
+import { Polly, Request, Response } from '@pollyjs/core'
 import { PuppeteerAdapter } from './polly/PuppeteerAdapter'
`

	expectedRanges := protocol.Ranges{{
		Start: protocol.Location{Line: 3, Column: 10, Offset: 200},
		End:   protocol.Location{Line: 3, Column: 15, Offset: 205},
	}, {
		Start: protocol.Location{Line: 3, Column: 25, Offset: 215},
		End:   protocol.Location{Line: 3, Column: 30, Offset: 220},
	}, {
		Start: protocol.Location{Line: 4, Column: 10, Offset: 239},
		End:   protocol.Location{Line: 4, Column: 15, Offset: 244},
	}, {
		Start: protocol.Location{Line: 4, Column: 44, Offset: 273},
		End:   protocol.Location{Line: 4, Column: 49, Offset: 278},
	}}

	require.Equal(t, expectedFormatted, formatted)
	require.Equal(t, expectedRanges, ranges)

}

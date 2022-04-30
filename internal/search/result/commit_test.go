package result

import (
	"testing"
	"testing/quick"

	"github.com/hexops/autogold"
)

func TestCommitSearchResult_Limit(t *testing.T) {
	f := func(nHighlights []int, limitInput uint32) bool {
		cr := &CommitMatch{
			MessagePreview: &MatchedString{
				MatchedRanges: make([]Range, len(nHighlights)),
			},
		}

		// It isn't interesting to test limit > ResultCount, so we bound it to
		// [1, ResultCount]
		count := cr.ResultCount()
		limit := (int(limitInput) % count) + 1

		after := cr.Limit(limit)
		newCount := cr.ResultCount()

		if after == 0 && newCount == limit {
			return true
		}

		t.Logf("failed limit=%d count=%d => after=%d newCount=%d", limit, count, after, newCount)
		return false
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error("quick check failed")
	}

	for nSymbols := 0; nSymbols <= 3; nSymbols++ {
		for limit := 0; limit <= nSymbols; limit++ {
			if !f(make([]int, nSymbols), uint32(limit)) {
				t.Error("small exhaustive check failed")
			}
		}
	}
}

func TestParseDiffString(t *testing.T) {
	input := `client/web/src/enterprise/codeintel/badge/components/IndexerSummary.module.scss client/web/src/enterprise/codeintel/badge/components/IndexerSummary.module.scss
@@ -1,2 +1,6 @@ ... +1
+.badge-wrapper {
+    font-size: 0.75rem;
+}
+
 .telemetric-redirect {
-    display: inline !important;
+    font-size: 0.75rem;
client/web/src/enterprise/codeintel/badge/components/IndexerSummary.tsx client/web/src/enterprise/codeintel/badge/components/IndexerSummary.tsx
@@ -57,3 +57,3 @@ export const IndexerSummary: React.FunctionComponent<IndexerSummaryProps> = ({
     return (
-        <div className="px-2 py-1">
+        <div className={classNames('px-2 py-1', styles.badgeWrapper)}>
             <div className="d-flex align-items-center">
@@ -61,3 +61,3 @@ export const IndexerSummary: React.FunctionComponent<IndexerSummaryProps> = ({
                     {summary.uploads.length + summary.indexes.length > 0 ? (
-                        <Badge variant="success" className={className}>
+                        <Badge variant="success" small={true} className={className}>
                             Enabled
@@ -65,3 +65,3 @@ export const IndexerSummary: React.FunctionComponent<IndexerSummaryProps> = ({
                     ) : summary.indexer?.url ? (
-                        <Badge variant="secondary" className={className}>
+                        <Badge variant="secondary" small={true} className={className}>
                             Configurable
client/web/src/enterprise/codeintel/badge/components/RequestLink.module.scss client/web/src/enterprise/codeintel/badge/components/RequestLink.module.scss
@@ -1,3 +1,5 @@ 
 .language-request {
+    font-size: 0.75rem;
     font-weight: normal;
+    display: inline !important;
 }
`

	res, err := parseDiffString(input)
	if err != nil {
		panic(err)
	}
	autogold.Equal(t, res)
}

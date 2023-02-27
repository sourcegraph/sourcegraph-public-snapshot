package result

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hexops/autogold/v2"
)

const input = `client/web/src/enterprise/codeintel/badge/components/IndexerSummary.module.scss client/web/src/enterprise/codeintel/badge/components/IndexerSummary.module.scss
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

func TestParseDiffString(t *testing.T) {
	res, err := ParseDiffString(input)
	if err != nil {
		panic(err)
	}
	autogold.ExpectFile(t, res)

	formatted := FormatDiffFiles(res)
	require.Equal(t, input, formatted)

}

func TestCommitDiffMatch(t *testing.T) {
	res, _ := ParseDiffString(input)
	commitDiff := &CommitDiffMatch{DiffFile: &res[0]}
	autogold.Expect("client/web/src/enterprise/codeintel/badge/components/IndexerSummary.module.scss").
		Equal(t, commitDiff.Path())
}

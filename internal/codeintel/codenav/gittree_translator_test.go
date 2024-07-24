package codenav

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	genslices "github.com/life4/genesis/slices"
	godiff "github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	sgtypes "github.com/sourcegraph/sourcegraph/internal/types"
)

var mockRepo sgtypes.Repo = sgtypes.Repo{ID: 50}

func rp(path string) core.RepoRelPath {
	return core.NewRepoRelPathUnchecked(path)
}

func diffMock(diff string) gitserver.Client {
	gs := gitserver.NewMockClient()
	gs.DiffFunc.SetDefaultHook(func(ctx context.Context, rn api.RepoName, do gitserver.DiffOptions) (*gitserver.DiffFileIterator, error) {
		return gitserver.NewDiffFileIterator(io.NopCloser(bytes.NewReader([]byte(diff)))), nil
	})
	return gs
}

func TestGetTargetCommitPositionFromSourcePosition(t *testing.T) {
	client := diffMock(hugoDiff)
	posIn := scip.Position{Line: 302, Character: 15}

	adjuster := NewGitTreeTranslator(client, mockRepo)
	posOutOpt, err := adjuster.TranslatePosition(context.Background(), "deadbeef1", "deadbeef2", rp("resources/image.go"), posIn)

	require.NoError(t, err)
	posOut, ok := posOutOpt.Get()
	require.Truef(t, ok, "expected translation to succeed")
	expectedPos := scip.Position{Line: 294, Character: 15}
	if diff := cmp.Diff(expectedPos, posOut); diff != "" {
		t.Errorf("unexpected position (-want +got):\n%s", diff)
	}
}

func TestGetTargetCommitPositionFromSourcePositionEmptyDiff(t *testing.T) {
	client := diffMock("")
	posIn := scip.Position{Line: 10, Character: 15}

	adjuster := NewGitTreeTranslator(client, mockRepo)
	posOutOpt, err := adjuster.TranslatePosition(context.Background(), "deadbeef1", "deadbeef2", rp("resources/image.go"), posIn)

	require.NoError(t, err)
	posOut, ok := posOutOpt.Get()
	require.Truef(t, ok, "expected translation to succeed")
	if diff := cmp.Diff(posOut, posIn); diff != "" {
		t.Errorf("unexpected position (-want +got):\n%s", diff)
	}
}

func TestGetTargetCommitPositionFromSourcePositionReverse(t *testing.T) {
	client := diffMock(hugoDiff)
	posIn := scip.Position{Line: 302, Character: 15}

	adjuster := NewGitTreeTranslator(client, mockRepo)
	posOutOpt, err := adjuster.TranslatePosition(context.Background(), "deadbeef2", "deadbeef1", rp("resources/image.go"), posIn)

	require.NoError(t, err)
	posOut, ok := posOutOpt.Get()
	require.Truef(t, ok, "expected translation to succeed")
	expectedPos := scip.Position{Line: 294, Character: 15}
	if diff := cmp.Diff(expectedPos, posOut); diff != "" {
		t.Errorf("unexpected position (-want +got):\n%s", diff)
	}
}

func TestGetTargetCommitRangeFromSourceRange(t *testing.T) {
	client := diffMock(hugoDiff)
	rIn := scip.Range{
		Start: scip.Position{Line: 302, Character: 15},
		End:   scip.Position{Line: 305, Character: 20},
	}

	adjuster := NewGitTreeTranslator(client, mockRepo)
	rOutOpt, err := adjuster.TranslateRange(context.Background(), "deadbeef1", "deadbeef2", rp("resources/image.go"), rIn)

	require.NoError(t, err)
	rOut, ok := rOutOpt.Get()
	require.Truef(t, ok, "expected translation to succeed")
	expectedRange := scip.Range{
		Start: scip.Position{Line: 294, Character: 15},
		End:   scip.Position{Line: 297, Character: 20},
	}
	if diff := cmp.Diff(expectedRange, rOut); diff != "" {
		t.Errorf("unexpected position (-want +got):\n%s", diff)
	}
}

func TestGetTargetCommitRangeFromSourceRangeEmptyDiff(t *testing.T) {
	client := diffMock("")
	rIn := scip.Range{
		Start: scip.Position{Line: 302, Character: 15},
		End:   scip.Position{Line: 305, Character: 20},
	}

	adjuster := NewGitTreeTranslator(client, mockRepo)
	rOutOpt, err := adjuster.TranslateRange(context.Background(), "deadbeef1", "deadbeef2", rp("resources/image.go"), rIn)

	require.NoError(t, err)
	rOut, ok := rOutOpt.Get()
	require.Truef(t, ok, "expected translation to succeed")
	if diff := cmp.Diff(rOut, rIn); diff != "" {
		t.Errorf("unexpected position (-want +got):\n%s", diff)
	}
}

func TestGetTargetCommitRangeFromSourceRangeReverse(t *testing.T) {
	client := diffMock(hugoDiff)
	rIn := scip.Range{
		Start: scip.Position{Line: 302, Character: 15},
		End:   scip.Position{Line: 305, Character: 20},
	}

	adjuster := NewGitTreeTranslator(client, mockRepo)
	rOutOpt, err := adjuster.TranslateRange(context.Background(), "deadbeef2", "deadbeef1", rp("resources/image.go"), rIn)

	require.NoError(t, err)
	rOut, ok := rOutOpt.Get()
	require.Truef(t, ok, "expected translation to succeed")
	expectedRange := scip.Range{
		Start: scip.Position{Line: 294, Character: 15},
		End:   scip.Position{Line: 297, Character: 20},
	}
	if diff := cmp.Diff(expectedRange, rOut); diff != "" {
		t.Errorf("unexpected position (-want +got):\n%s", diff)
	}
}

type gitTreeTranslatorTestCase struct {
	diff         string // The git diff output
	diffName     string // The git diff output name
	description  string // The description of the test
	line         int32  // The target line (one-indexed)
	expectedOk   bool   // Whether the operation should succeed
	expectedLine int32  // The expected adjusted line (one-indexed)
}

// hugoDiff is a diff from github.com/gohugoio/hugo generated via the following command.
// git diff-tree --patch --find-renames --full-index --inter-hunk-context=0 --unified=0 --no-prefix 8947c3fa0beec021e14b3f8040857335e1ecd473 3e9db2ad951dbb1000cd0f8f25e4a95445046679 -- resources/image.go
const hugoDiff = `
diff --git resources/image.go resources/image.go
index d1d9f650d673e35359444dc9df4f1e24e2cd4fbc..076f2ae4d63b1b6e2de1e3308f6e7bdb791d4d33 100644
--- resources/image.go
+++ resources/image.go
@@ -39 +38,0 @@ import (
-	"github.com/pkg/errors"
@@ -238 +237 @@ func (i *imageResource) doWithImageConfig(conf images.ImageConfig, f func(src im
-	img, err := i.getSpec().imageCache.getOrCreate(i, conf, func() (*imageResource, image.Image, error) {
+	return i.getSpec().imageCache.getOrCreate(i, conf, func() (*imageResource, image.Image, error) {
@@ -295,7 +293,0 @@ func (i *imageResource) doWithImageConfig(conf images.ImageConfig, f func(src im
-
-	if err != nil {
-		if i.root != nil && i.root.getFileInfo() != nil {
-			return nil, errors.Wrapf(err, "image %q", i.root.getFileInfo().Meta().Filename())
-		}
-	}
-	return img, nil
`

var hugoTestCases = []gitTreeTranslatorTestCase{
	// Between hunks
	{hugoDiff, "hugo", "before first hunk", 10, true, 10},
	{hugoDiff, "hugo", "between hunks (1x deletion)", 150, true, 149},
	{hugoDiff, "hugo", "between hunks (1x deletion, 1x edit)", 250, true, 249},
	{hugoDiff, "hugo", "after last hunk (2x deletions, 1x edit)", 350, true, 342},

	// Hunk 1
	{hugoDiff, "hugo", "before first hunk deletion", 38, true, 38},
	{hugoDiff, "hugo", "on first hunk deletion", 39, false, 0},
	{hugoDiff, "hugo", "after first hunk deletion", 40, true, 39},

	// Hunk 1 (lower border)
	{hugoDiff, "hugo", "inside first hunk context (last line)", 43, true, 42},
	{hugoDiff, "hugo", "directly after first hunk", 44, true, 43},

	// Hunk 2
	{hugoDiff, "hugo", "before second hunk edit", 237, true, 236},
	{hugoDiff, "hugo", "on edited hunk edit", 238, false, 0},
	{hugoDiff, "hugo", "after second hunk edit", 239, true, 238},

	// Hunk 3
	{hugoDiff, "hugo", "before third hunk deletion", 294, true, 293},
	{hugoDiff, "hugo", "on third hunk deletion", 295, false, 0},
	{hugoDiff, "hugo", "on third hunk deletion", 301, false, 0},
	{hugoDiff, "hugo", "after third hunk deletion", 302, true, 294},
}

// prometheusDiff is a diff from github.com/prometheus/prometheus generated via the following command.
// git diff-tree --patch --find-renames --full-index --inter-hunk-context=0 --unified=0 --no-prefix 52025bd7a9446c3178bf01dd2949d4874dd45f24 45fbed94d6ee17840254e78cfc421ab1db78f734 -- discovery/manager.go
const prometheusDiff = `
diff --git discovery/manager.go discovery/manager.go
index 49bcbf86b7baa70bff34b0fa306ca20877f5640e..d135cd54e700ea67963a186ca370d59466f9eb78 100644
--- discovery/manager.go
+++ discovery/manager.go
@@ -296,3 +295,0 @@ func (m *Manager) updateGroup(poolKey poolKey, tgs []*targetgroup.Group) {
-	if _, ok := m.targets[poolKey]; !ok {
-		m.targets[poolKey] = make(map[string]*targetgroup.Group)
-	}
@@ -300,0 +298,3 @@ func (m *Manager) updateGroup(poolKey poolKey, tgs []*targetgroup.Group) {
+			if _, ok := m.targets[poolKey]; !ok {
+				m.targets[poolKey] = make(map[string]*targetgroup.Group)
+			}
`

var prometheusTestCases = []gitTreeTranslatorTestCase{
	{prometheusDiff, "prometheus", "before hunk", 100, true, 100},
	{prometheusDiff, "prometheus", "before deletion", 295, true, 295},
	{prometheusDiff, "prometheus", "on deletion 1", 296, false, 0},
	{prometheusDiff, "prometheus", "on deletion 2", 297, false, 0},
	{prometheusDiff, "prometheus", "on deletion 3", 298, false, 0},
	{prometheusDiff, "prometheus", "after deletion", 299, true, 296},
	{prometheusDiff, "prometheus", "before insertion", 300, true, 297},
	{prometheusDiff, "prometheus", "after insertion", 301, true, 301},
	{prometheusDiff, "prometheus", "after hunk", 500, true, 500},
}

func TestRawGetTargetCommitPositionFromSourcePosition(t *testing.T) {
	for _, testCase := range append(append([]gitTreeTranslatorTestCase(nil), hugoTestCases...), prometheusTestCases...) {
		name := fmt.Sprintf("%s : %s", testCase.diffName, testCase.description)

		t.Run(name, func(t *testing.T) {
			diff, err := godiff.NewFileDiffReader(bytes.NewReader([]byte(testCase.diff))).Read()
			require.NoError(t, err, "unexpected error reading file diff")
			hunks := genslices.Map(diff.Hunks, newCompactHunk)

			pos := scip.Position{
				Line:      testCase.line - 1, // 1-index -> 0-index
				Character: 10,
			}

			if adjusted, ok := translatePosition(hunks, pos).Get(); ok != testCase.expectedOk {
				t.Errorf("unexpected ok. want=%v have=%v", testCase.expectedOk, ok)
			} else if ok {
				// Adjust from zero-index to one-index
				if adjusted.Line+1 != testCase.expectedLine {
					t.Errorf("unexpected line. want=%d have=%d", testCase.expectedLine, adjusted.Line+1) // 0-index -> 1-index
				}
				if adjusted.Character != 10 {
					t.Errorf("unexpected character. want=%d have=%d", 10, adjusted.Character)
				}
			}
		})
	}
}

package codenav

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	godiff "github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/sourcegraph/internal/api"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	sgtypes "github.com/sourcegraph/sourcegraph/internal/types"
)

func TestGetTargetCommitPathFromSourcePath(t *testing.T) {
	client := gitserver.NewMockClient()

	args := &requestArgs{
		repo:   &sgtypes.Repo{ID: 50},
		commit: "deadbeef1",
		path:   "/foo/bar.go",
	}
	adjuster := NewGitTreeTranslator(client, args, nil)
	path, ok, err := adjuster.GetTargetCommitPathFromSourcePath(context.Background(), "deadbeef2", "/foo/bar.go", false)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if !ok {
		t.Errorf("expected translation to succeed")
	}
	if path != "/foo/bar.go" {
		t.Errorf("unexpected path. want=%s have=%s", "/foo/bar.go", path)
	}
}

func TestGetTargetCommitPositionFromSourcePosition(t *testing.T) {
	client := gitserver.NewMockClientWithExecReader(nil, func(_ context.Context, _ api.RepoName, args []string) (reader io.ReadCloser, err error) {
		expectedArgs := []string{"diff", "deadbeef1", "deadbeef2", "--", "/foo/bar.go"}
		if diff := cmp.Diff(expectedArgs, args); diff != "" {
			t.Errorf("unexpected exec reader args (-want +got):\n%s", diff)
		}

		return io.NopCloser(bytes.NewReader([]byte(hugoDiff))), nil
	})

	posIn := shared.Position{Line: 302, Character: 15}

	args := &requestArgs{
		repo:   &sgtypes.Repo{ID: 50},
		commit: "deadbeef1",
		path:   "/foo/bar.go",
	}
	adjuster := NewGitTreeTranslator(client, args, nil)
	path, posOut, ok, err := adjuster.GetTargetCommitPositionFromSourcePosition(context.Background(), "deadbeef2", posIn, false)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if !ok {
		t.Errorf("expected translation to succeed")
	}
	if path != "/foo/bar.go" {
		t.Errorf("unexpected path. want=%s have=%s", "/foo/bar.go", path)
	}

	expectedPos := shared.Position{Line: 294, Character: 15}
	if diff := cmp.Diff(expectedPos, posOut); diff != "" {
		t.Errorf("unexpected position (-want +got):\n%s", diff)
	}
}

func TestGetTargetCommitPositionFromSourcePositionEmptyDiff(t *testing.T) {
	client := gitserver.NewMockClientWithExecReader(nil, func(_ context.Context, _ api.RepoName, args []string) (reader io.ReadCloser, err error) {
		return io.NopCloser(bytes.NewReader(nil)), nil
	})

	posIn := shared.Position{Line: 10, Character: 15}

	args := &requestArgs{
		repo:   &sgtypes.Repo{ID: 50},
		commit: "deadbeef1",
		path:   "/foo/bar.go",
	}
	adjuster := NewGitTreeTranslator(client, args, nil)
	path, posOut, ok, err := adjuster.GetTargetCommitPositionFromSourcePosition(context.Background(), "deadbeef2", posIn, false)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if !ok {
		t.Errorf("expected translation to succeed")
	}
	if path != "/foo/bar.go" {
		t.Errorf("unexpected path. want=%s have=%s", "/foo/bar.go", path)
	}
	if diff := cmp.Diff(posOut, posIn); diff != "" {
		t.Errorf("unexpected position (-want +got):\n%s", diff)
	}
}

func TestGetTargetCommitPositionFromSourcePositionReverse(t *testing.T) {
	client := gitserver.NewMockClientWithExecReader(nil, func(_ context.Context, _ api.RepoName, args []string) (reader io.ReadCloser, err error) {
		expectedArgs := []string{"diff", "deadbeef2", "deadbeef1", "--", "/foo/bar.go"}
		if diff := cmp.Diff(expectedArgs, args); diff != "" {
			t.Errorf("unexpected exec reader args (-want +got):\n%s", diff)
		}

		return io.NopCloser(bytes.NewReader([]byte(hugoDiff))), nil
	})

	posIn := shared.Position{Line: 302, Character: 15}

	args := &requestArgs{
		repo:   &sgtypes.Repo{ID: 50},
		commit: "deadbeef1",
		path:   "/foo/bar.go",
	}
	adjuster := NewGitTreeTranslator(client, args, nil)
	path, posOut, ok, err := adjuster.GetTargetCommitPositionFromSourcePosition(context.Background(), "deadbeef2", posIn, true)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if !ok {
		t.Errorf("expected translation to succeed")
	}
	if path != "/foo/bar.go" {
		t.Errorf("unexpected path. want=%s have=%s", "/foo/bar.go", path)
	}

	expectedPos := shared.Position{Line: 294, Character: 15}
	if diff := cmp.Diff(expectedPos, posOut); diff != "" {
		t.Errorf("unexpected position (-want +got):\n%s", diff)
	}
}

func TestGetTargetCommitRangeFromSourceRange(t *testing.T) {
	client := gitserver.NewMockClientWithExecReader(nil, func(_ context.Context, _ api.RepoName, args []string) (reader io.ReadCloser, err error) {
		expectedArgs := []string{"diff", "deadbeef1", "deadbeef2", "--", "/foo/bar.go"}
		if diff := cmp.Diff(expectedArgs, args); diff != "" {
			t.Errorf("unexpected exec reader args (-want +got):\n%s", diff)
		}

		return io.NopCloser(bytes.NewReader([]byte(hugoDiff))), nil
	})

	rIn := shared.Range{
		Start: shared.Position{Line: 302, Character: 15},
		End:   shared.Position{Line: 305, Character: 20},
	}

	args := &requestArgs{
		repo:   &sgtypes.Repo{ID: 50},
		commit: "deadbeef1",
		path:   "/foo/bar.go",
	}
	adjuster := NewGitTreeTranslator(client, args, nil)
	path, rOut, ok, err := adjuster.GetTargetCommitRangeFromSourceRange(context.Background(), "deadbeef2", "/foo/bar.go", rIn, false)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if !ok {
		t.Errorf("expected translation to succeed")
	}
	if path != "/foo/bar.go" {
		t.Errorf("unexpected path. want=%s have=%s", "/foo/bar.go", path)
	}

	expectedRange := shared.Range{
		Start: shared.Position{Line: 294, Character: 15},
		End:   shared.Position{Line: 297, Character: 20},
	}
	if diff := cmp.Diff(expectedRange, rOut); diff != "" {
		t.Errorf("unexpected position (-want +got):\n%s", diff)
	}
}

func TestGetTargetCommitRangeFromSourceRangeEmptyDiff(t *testing.T) {
	client := gitserver.NewMockClientWithExecReader(nil, func(_ context.Context, _ api.RepoName, args []string) (reader io.ReadCloser, err error) {
		return io.NopCloser(bytes.NewReader([]byte(nil))), nil
	})

	rIn := shared.Range{
		Start: shared.Position{Line: 302, Character: 15},
		End:   shared.Position{Line: 305, Character: 20},
	}

	args := &requestArgs{
		repo:   &sgtypes.Repo{ID: 50},
		commit: "deadbeef1",
		path:   "/foo/bar.go",
	}
	adjuster := NewGitTreeTranslator(client, args, nil)
	path, rOut, ok, err := adjuster.GetTargetCommitRangeFromSourceRange(context.Background(), "deadbeef2", "/foo/bar.go", rIn, false)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if !ok {
		t.Errorf("expected translation to succeed")
	}
	if path != "/foo/bar.go" {
		t.Errorf("unexpected path. want=%s have=%s", "/foo/bar.go", path)
	}
	if diff := cmp.Diff(rOut, rIn); diff != "" {
		t.Errorf("unexpected position (-want +got):\n%s", diff)
	}
}

func TestGetTargetCommitRangeFromSourceRangeReverse(t *testing.T) {
	client := gitserver.NewMockClientWithExecReader(nil, func(_ context.Context, _ api.RepoName, args []string) (reader io.ReadCloser, err error) {
		expectedArgs := []string{"diff", "deadbeef2", "deadbeef1", "--", "/foo/bar.go"}
		if diff := cmp.Diff(expectedArgs, args); diff != "" {
			t.Errorf("unexpected exec reader args (-want +got):\n%s", diff)
		}

		return io.NopCloser(bytes.NewReader([]byte(hugoDiff))), nil
	})

	rIn := shared.Range{
		Start: shared.Position{Line: 302, Character: 15},
		End:   shared.Position{Line: 305, Character: 20},
	}

	args := &requestArgs{
		repo:   &sgtypes.Repo{ID: 50},
		commit: "deadbeef1",
		path:   "/foo/bar.go",
	}
	adjuster := NewGitTreeTranslator(client, args, nil)
	path, rOut, ok, err := adjuster.GetTargetCommitRangeFromSourceRange(context.Background(), "deadbeef2", "/foo/bar.go", rIn, true)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if !ok {
		t.Errorf("expected translation to succeed")
	}
	if path != "/foo/bar.go" {
		t.Errorf("unexpected path. want=%s have=%s", "/foo/bar.go", path)
	}

	expectedRange := shared.Range{
		Start: shared.Position{Line: 294, Character: 15},
		End:   shared.Position{Line: 297, Character: 20},
	}
	if diff := cmp.Diff(expectedRange, rOut); diff != "" {
		t.Errorf("unexpected position (-want +got):\n%s", diff)
	}
}

type gitTreeTranslatorTestCase struct {
	diff         string // The git diff output
	diffName     string // The git diff output name
	description  string // The description of the test
	line         int    // The target line (one-indexed)
	expectedOk   bool   // Whether the operation should succeed
	expectedLine int    // The expected adjusted line (one-indexed)
}

// hugoDiff is a diff from github.com/gohugoio/hugo generated via the following command.
// git diff 8947c3fa0beec021e14b3f8040857335e1ecd473 3e9db2ad951dbb1000cd0f8f25e4a95445046679 -- resources/image.go
const hugoDiff = `
diff --git a/resources/image.go b/resources/image.go
index d1d9f650d673..076f2ae4d63b 100644
--- a/resources/image.go
+++ b/resources/image.go
@@ -36,7 +36,6 @@ import (

        "github.com/gohugoio/hugo/resources/resource"

-       "github.com/sourcegraph/sourcegraph/lib/errors"
        _errors "github.com/sourcegraph/sourcegraph/lib/errors"

        "github.com/gohugoio/hugo/helpers"
@@ -235,7 +234,7 @@ const imageProcWorkers = 1
 var imageProcSem = make(chan bool, imageProcWorkers)

 func (i *imageResource) doWithImageConfig(conf images.ImageConfig, f func(src image.Image) (image.Image, error)) (resource.Image, error) {
-       img, err := i.getSpec().imageCache.getOrCreate(i, conf, func() (*imageResource, image.Image, error) {
+       return i.getSpec().imageCache.getOrCreate(i, conf, func() (*imageResource, image.Image, error) {
                imageProcSem <- true
                defer func() {
                        <-imageProcSem
@@ -292,13 +291,6 @@ func (i *imageResource) doWithImageConfig(conf images.ImageConfig, f func(src im

                return ci, converted, nil
        })
-
-       if err != nil {
-               if i.root != nil && i.root.getFileInfo() != nil {
-                       return nil, errors.Wrapf(err, "image %q", i.root.getFileInfo().Meta().Filename())
-               }
-       }
-       return img, nil
 }

 func (i *imageResource) decodeImageConfig(action, spec string) (images.ImageConfig, error) {
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
// git diff 52025bd7a9446c3178bf01dd2949d4874dd45f24 45fbed94d6ee17840254e78cfc421ab1db78f734 -- discovery/manager.go
const prometheusDiff = `
diff --git a/discovery/manager.go b/discovery/manager.go
index 49bcbf86b7ba..d135cd54e700 100644
--- a/discovery/manager.go
+++ b/discovery/manager.go
@@ -293,11 +293,11 @@ func (m *Manager) updateGroup(poolKey poolKey, tgs []*targetgroup.Group) {
        m.mtx.Lock()
        defer m.mtx.Unlock()

-       if _, ok := m.targets[poolKey]; !ok {
-               m.targets[poolKey] = make(map[string]*targetgroup.Group)
-       }
        for _, tg := range tgs {
                if tg != nil { // Some Discoverers send nil target group so need to check for it to avoid panics.
+                       if _, ok := m.targets[poolKey]; !ok {
+                               m.targets[poolKey] = make(map[string]*targetgroup.Group)
+                       }
                        m.targets[poolKey][tg.Source] = tg
                }
        }
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
			if err != nil {
				t.Fatalf("unexpected error reading file diff: %s", err)
			}
			hunks := diff.Hunks

			pos := shared.Position{
				Line:      testCase.line - 1, // 1-index -> 0-index
				Character: 10,
			}

			if adjusted, ok := translatePosition(hunks, pos); ok != testCase.expectedOk {
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

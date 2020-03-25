package resolvers

import (
	"testing"

	"github.com/sourcegraph/go-lsp"
)

// hugoDiff is a diff from github.com/gohugoio/hugo generated via the following command.
// git diff 8947c3fa0beec021e14b3f8040857335e1ecd473 3e9db2ad951dbb1000cd0f8f25e4a95445046679 -- resources/image.go
const hugoDiff = `
diff --git a/resources/image.go b/resources/image.go
index d1d9f650d673..076f2ae4d63b 100644
--- a/resources/image.go
+++ b/resources/image.go
@@ -36,7 +36,6 @@ import (

        "github.com/gohugoio/hugo/resources/resource"

-       "github.com/pkg/errors"
        _errors "github.com/pkg/errors"

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

func TestAdjustPositionFromDiff(t *testing.T) {
	testCases := []struct {
		diff         string // The git diff output
		line         int    // The target line (one-indexed)
		expectedLine *int   // The expected adjusted line (one-indexed)
	}{
		// Between hunks
		{hugoDiff, 10, intPtr(10)},   // before first hunk
		{hugoDiff, 150, intPtr(149)}, // between hunks (1x deletion)
		{hugoDiff, 250, intPtr(249)}, // between hunks (1x deletion, 1x edit)
		{hugoDiff, 350, intPtr(342)}, // after last hunk (2x deletions, 1x edit)

		// Hunk 1
		{hugoDiff, 38, intPtr(38)}, // before first hunk deletion
		{hugoDiff, 39, nil},        // on first hunk deletion
		{hugoDiff, 40, intPtr(39)}, // after first hunk deletion

		// Hunk 1 (lower border)
		{hugoDiff, 43, intPtr(42)}, // inside first hunk context (last line)
		{hugoDiff, 44, intPtr(43)}, // directly after first hunk

		// Hunk 2
		{hugoDiff, 237, intPtr(236)}, // before second hunk edit
		{hugoDiff, 238, nil},         // on edited hunk edit
		{hugoDiff, 239, intPtr(238)}, // after second hunk edit

		// Hunk 3
		{hugoDiff, 294, intPtr(293)}, // before third hunk deletion
		{hugoDiff, 295, nil},         // on third hunk deletion
		{hugoDiff, 301, nil},         // on third hunk deletion
		{hugoDiff, 302, intPtr(294)}, // after third hunk deletion

		// Prometheus
		{prometheusDiff, 100, intPtr(100)}, // before hunk
		{prometheusDiff, 295, intPtr(295)}, // before deletion
		{prometheusDiff, 296, nil},         // on deletion
		{prometheusDiff, 297, nil},         // on deletion
		{prometheusDiff, 298, nil},         // on deletion
		{prometheusDiff, 299, intPtr(296)}, // after deletion
		{prometheusDiff, 300, intPtr(297)}, // before insertion
		{prometheusDiff, 301, intPtr(301)}, // after insertion
		{prometheusDiff, 500, intPtr(500)}, // after hunk
	}

	for i, testCase := range testCases {
		adjuster, err := newPositionAdjusterFromDiffOutput([]byte(testCase.diff))
		if err != nil {
			t.Errorf("Unexpected error in test case #%d: %s", i+1, err)
		}

		// Adjust from one-index to zero-index
		pos := lsp.Position{Line: testCase.line - 1, Character: 10}
		adjusted := adjuster.adjustPosition(pos)

		if testCase.expectedLine == nil {
			if adjusted != nil {
				t.Errorf("Unexpected line in test case #%d: got %d expected nil", i+1, adjusted.Line)
			}
		} else if adjusted == nil {
			t.Errorf("Unexpected nil in test case #%d: expected %d", i+1, *testCase.expectedLine)
		} else {
			// Adjust from zero-index to one-index
			if adjusted.Line+1 != *testCase.expectedLine {
				t.Errorf("Unexpected line in test case #%d: got %d expected %d", i+1, adjusted.Line, *testCase.expectedLine)
			}
			if adjusted.Character != 10 {
				t.Errorf("Unexpected character in test case #%d: %d", i+1, adjusted.Character)
			}
		}
	}
}

func intPtr(v int) *int {
	return &v
}

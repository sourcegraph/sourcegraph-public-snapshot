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

func TestAdjustPosition(t *testing.T) {
	testCases := []struct {
		diff         string // The git diff output
		diffName     string // The git diff output name
		description  string // The description of the test
		line         int    // The target line (one-indexed)
		expectedOk   bool   // Whether the operation should succeed
		expectedLine int    // The expected adjusted line (one-indexed)
	}{
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

		// Prometheus
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

	for _, testCase := range testCases {
		adjuster, err := newPositionAdjusterFromDiffOutput([]byte(testCase.diff))
		if err != nil {
			t.Errorf("Unexpected error in test case %s::%s: %s", testCase.diffName, testCase.description, err)
		}

		// Adjust from one-index to zero-index
		pos := lsp.Position{Line: testCase.line - 1, Character: 10}
		adjusted, ok := adjuster.adjustPosition(pos)

		if ok != testCase.expectedOk {
			t.Errorf("Test %s::%s: got %v expected %v", testCase.diffName, testCase.description, ok, testCase.expectedOk)
		} else if ok {
			// Adjust from zero-index to one-index
			if adjusted.Line+1 != testCase.expectedLine {
				t.Errorf("Test %s::%s: got %d expected %d", testCase.diffName, testCase.description, adjusted.Line+1, testCase.expectedLine)
			}
			if adjusted.Character != 10 {
				t.Errorf("Test %s::%s: got %d expected %d", testCase.diffName, testCase.description, adjusted.Character, 10)
			}
		}
	}
}

func TestAdjustPositionEmptyDiff(t *testing.T) {
	adjuster, err := newPositionAdjusterFromDiffOutput(nil)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	adjusted, ok := adjuster.adjustPosition(lsp.Position{Line: 25, Character: 10})
	if !ok {
		t.Errorf("Unexpected failure adjusting position")
	}
	if adjusted.Line != 25 || adjusted.Character != 10 {
		t.Errorf("Unexpected result: got %d:%d expected %d:%d", adjusted.Line, adjusted.Character, 25, 10)
	}
}

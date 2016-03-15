// +build uitest

package ui

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"

	"golang.org/x/net/context"
)

var (
	buildBundleJSErr  error
	buildBundleJSOnce sync.Once
)

func buildBundleJS() {
	log.Println("Building bundle.js for React component rendering tests. This could take a while.")
	cmd := exec.Command("npm", "run", "build")
	cmd.Dir = "../../" // app/ dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error building bundle.js for React component rendering tests (%s). Output follows:\n\n%s\n\n", err, out)
		buildBundleJSErr = err
	}
}

func ensureBundleJSFound(t *testing.T) {
	if _, _, err := getBundleJS(); os.IsNotExist(err) {
		// This could use a stale (but existent) bundle.js, but that
		// is a tradeoff we're willing to make to avoid having to
		// rebuild bundle.js each time you run the tests. CI will
		// build it fresh, so we have a safety check.
		buildBundleJSOnce.Do(buildBundleJS)
		if buildBundleJSErr != nil {
			t.Fatalf("React component rendering test requires bundle.js, which failed to build: %s", err)
		}
	} else if err != nil {
		t.Fatal(err)
	}
}

func TestRenderReactComponent_BuildIndicatorContainer(t *testing.T) {
	ensureBundleJSFound(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	resp, err := renderReactComponent(ctx, "sourcegraph/build/BuildIndicatorContainer", map[string]interface{}{
		"repo":      "myrepo",
		"commitID":  "master",
		"buildable": true,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if want := `<a title="Build `; !strings.HasPrefix(resp, want) {
		t.Errorf("got %q, want it to have prefix %q", resp, want)
	}
}

func TestRenderReactComponent_BlobRouter(t *testing.T) {
	ensureBundleJSFound(t)
	entrySpec := sourcegraph.TreeEntrySpec{
		RepoRev: sourcegraph.RepoRevSpec{
			RepoSpec: sourcegraph.RepoSpec{URI: "myrepo"},
			Rev:      "master",
			CommitID: "c",
		},
		Path: "myfile.txt",
	}

	var stores StoreData
	stores.BlobStore.AddFile(entrySpec, &sourcegraph.TreeEntry{ContentsString: "abcdefghi\njklmnopqr\nst"})
	stores.BlobStore.AddAnnotations(
		&sourcegraph.AnnotationsListOptions{
			Entry: entrySpec,
			Range: &sourcegraph.FileRange{},
		},
		&sourcegraph.AnnotationList{
			Annotations: []*sourcegraph.Annotation{
				{URL: "a", StartByte: 1, EndByte: 5},

				// Multiple URLs
				{URL: "b", StartByte: 3, EndByte: 6},
				{URL: "b2", StartByte: 3, EndByte: 6},
				{Class: "b", StartByte: 3, EndByte: 6},

				{URL: "c", StartByte: 6, EndByte: 12},
				{Class: "c", StartByte: 6, EndByte: 12},
				{URL: "d", StartByte: 13, EndByte: 15},
				{Class: "a", StartByte: 1, EndByte: 5},
				{Class: "b", StartByte: 15, EndByte: 18},
				{Class: "c", StartByte: 18, EndByte: 19},
				{Class: "d", StartByte: 2, EndByte: 4},
			},
			LineStartBytes: []uint32{0, 10, 20},
		},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	resp, err := renderReactComponent(ctx, "sourcegraph/LocationAdaptor", map[string]interface{}{
		"component": "sourcegraph/blob/BlobRouter",
		"location":  "/myrepo@master/.tree/myfile.txt",
	}, &stores)
	if err != nil {
		t.Fatal(err)
	}

	if want := `<a class="ref" href="a"`; !strings.Contains(resp, want) {
		t.Errorf("got %q, want it to contain %q", resp, want)
	}
}

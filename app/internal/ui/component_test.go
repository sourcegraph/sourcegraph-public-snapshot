// +build uitest

package ui

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

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

func TestRenderRouterState(t *testing.T) {
	ensureBundleJSFound(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := renderRouterState(ctx, &renderState{Location: "/login"})
	if err != nil {
		t.Fatal(err)
	}

	if want := `<div data-reactroot=`; !strings.Contains(string(resp), want) {
		t.Errorf("got %q, want it to contain %q", resp, want)
	}
}

func TestRenderRouterState_stress(t *testing.T) {
	ensureBundleJSFound(t)

	tmp := renderPoolSize
	renderPoolSize = 3
	defer func() {
		renderPoolSize = tmp
	}()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err := renderRouterState(ctx, &renderState{
				Location: fmt.Sprintf("/login#%d", i), // bypass render cache
			})
			if err != nil {
				t.Fatal(err)
			}
		}(i)
	}
	wg.Wait()
}

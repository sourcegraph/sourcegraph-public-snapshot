// +build exectest

package ui_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/server/testserver"
	"src.sourcegraph.com/sourcegraph/ui/payloads"
	"src.sourcegraph.com/sourcegraph/util/testutil"
)

// TODO(slimsag): If we start writing more UI tests, use a centralized
// testserver instance for better perf / lower test startup overhead.

// TestRepoTree_FileRange_lg tests that specifying RepoTreeGetOptions.GetFileOptions.FileRange
// fields as query parameters works. The frontend uses these to implement hunk
// expansion in diff views.
func TestRepoTree_FileRange_lg(t *testing.T) {
	if testserver.Store == "pgsql" {
		t.Skip("pgsql local store can only create mirror repos")
	}

	// Initialize a server instance.
	a, ctx := testserver.NewUnstartedServer()
	a.Config.ServeFlags = append(a.Config.ServeFlags,
		&authutil.Flags{DisableAccessControl: true},
	)
	if err := a.Start(); err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	// Create and push a repo with some files.
	files := map[string]string{
		"one": "first\nawesome\nfile\ncontents\n",
		"two": "second\nawesome\nfile\ncontents\n",
	}
	_, done, err := testutil.CreateAndPushRepoFiles(t, ctx, "r/r", files)
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	// Fetch two lines.
	resp, err := http.Get(a.Config.Serve.AppURL + ".ui/r/r/.tree/two?StartLine=1&EndLine=3")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Logf("Expected StatusCode == 200, Got %v\n", resp.StatusCode)
		t.Logf("Body: %q\n", string(body))
		t.FailNow()
	}

	// Verify the content type (or else it'll end up in the frontend as a string
	// instead of a JS object).
	wantEncoding := "application/json"
	if got := resp.Header.Get("Content-Type"); got != wantEncoding {
		t.Fatalf("Got Content-Type header %q want %q", got, wantEncoding)
	}

	// Verify the CodeFile.Entry is the one we asked for.
	var payload payloads.CodeFile
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Entry.FileRange.StartLine != 1 || payload.Entry.FileRange.EndLine != 3 {
		t.Fatalf("got unexpected StartLine:%v / EndLine:%v\n", payload.Entry.FileRange.StartLine, payload.Entry.FileRange.EndLine)
	}
}

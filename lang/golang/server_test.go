package golang

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

var updateFixtures = flag.Bool("fixtures.update", false, "update the expected files with actual")

var skipFixtures = map[string]string{
	"testdata/definition-external.json": "We do not support external symbol lookup yet",
}

func testFixtures(t *testing.T, h jsonrpc2.BatchHandler) {
	cases, err := filepath.Glob("testdata/*.json")
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cases {
		if reason, shouldSkip := skipFixtures[c]; shouldSkip {
			t.Logf("SKIP %s: %s", c, reason)
			continue
		}

		var req []*jsonrpc2.Request
		unmarshalFile(t, c, &req)

		// Test data specifies relative paths, but the language server expects
		// an absolute path. Make the path absolute now.
		var init lsp.InitializeParams
		err := json.Unmarshal(*req[0].Params, &init)
		if err != nil {
			t.Fatal(err)
		}
		init.RootPath, err = filepath.Abs(init.RootPath)
		if err != nil {
			t.Fatal(err)
		}
		req[0].SetParams(init)

		resp := h.HandleBatch(req)
		marshalFile(t, c+".actual", resp)
		if *updateFixtures {
			marshalFile(t, c+".expected", resp)
		}

		out, err := exec.Command("diff", c+".expected", c+".actual").Output()
		if err != nil {
			t.Errorf("unexpected response, output of diff %s %s:\n%s", c+".expected", c+".actual", string(out))
		}
	}
}

func unmarshalFile(t *testing.T, path string, v interface{}) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(b, v)
	if err != nil {
		t.Fatal(err)
	}
}

func marshalFile(t *testing.T, path string, v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(path, b, 0600)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFixtures(t *testing.T) {
	checkExecDeps(t)
	testFixtures(t, &Handler{})
}

func checkExecDeps(t *testing.T) {
	deps := map[string]string{
		"godef": "github.com/rogpeppe/godef",
		"guru":  "golang.org/x/tools/cmd/guru",
	}
	missing := []string{}
	for cmd, pkg := range deps {
		if _, err := exec.LookPath(cmd); err != nil {
			if os.Getenv("CI") == "" {
				t.Fatalf("Missing %s. Please run go get %s", cmd, pkg)
			}
			missing = append(missing, pkg)
		}
	}
	if len(missing) > 0 {
		t.Logf("go get %s", strings.Join(missing, " "))
		args := append([]string{"get"}, missing...)
		_, err := exec.Command("go", args...).Output()
		if err != nil {
			t.Fatal(err)
		}
	}
}

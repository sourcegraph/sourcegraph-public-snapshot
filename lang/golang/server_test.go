package golang

import (
	"context"
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

func TestFixtures(t *testing.T) {
	checkExecDeps(t)
	defer os.RemoveAll("testdata/cache")

	// enable gog for tests, even though disabled by default. This is to
	// prevent regressions in case we do want to use it.
	gogEnabled = true

	cases, err := filepath.Glob("testdata/*.json")
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cases {
		t.Run(filepath.Base(c), func(t *testing.T) {
			testFixture(t, c)
		})
	}
}

func testFixture(t *testing.T, c string) {
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
	if err := req[0].SetParams(init); err != nil {
		t.Fatal(err)
	}

	h := &Handler{}
	resps := make([]jsonrpc2.Response, len(req))
	for i, req := range req {
		result, err := h.Handle(context.Background(), nil, req)
		if err != nil {
			t.Errorf("call %v: %s", req, err)
			continue
		}
		resps[i].ID = uint64(i)
		if err := resps[i].SetResult(result); err != nil {
			t.Fatal(err)
		}
	}
	marshalFile(t, c+".actual", resps)
	if *updateFixtures {
		marshalFile(t, c+".expected", resps)
	}

	out, err := exec.Command("diff", c+".expected", c+".actual").Output()
	if err != nil {
		t.Errorf("unexpected response, output of diff %s %s:\n%s", c+".expected", c+".actual", string(out))
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

func checkExecDeps(t *testing.T) {
	deps := map[string]string{
		"godef": "github.com/sourcegraph/godef",
		"gog":   "sourcegraph.com/sourcegraph/srclib-go/gog/cmd/gog",
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

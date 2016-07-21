package golang

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
)

var updateFixtures = flag.Bool("fixtures.update", false, "update the expected files with actual")

func testFixtures(t *testing.T, h jsonrpc2.BatchHandler) {
	cases, err := filepath.Glob("testdata/*.json")
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cases {
		var req []*jsonrpc2.Request
		unmarshalFile(t, c, &req)

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
	testFixtures(t, &Handler{})
}

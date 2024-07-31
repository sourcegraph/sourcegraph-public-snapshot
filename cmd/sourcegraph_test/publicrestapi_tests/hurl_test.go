package publicrestapi_tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/sourcegraph_test"
)

func TestPublicRESTAPI(t *testing.T) {
	if os.Getenv("HTTP_RECORD_REPLAY_MODE") == "" {
		os.Setenv("HTTP_RECORD_REPLAY_MODE", "replay")
	}
	os.Setenv("HTTP_RECORD_REPLAY_DIR", filepath.Join("cmd/sourcegraph_test/publicrestapi_tests/recordings"))
	os.Setenv("HTTP_RECORD_REPLAY_NAME", "publicrestapi")

	go sourcegraph_test.RunSingleStaticBinaryMain()

	files, err := filepath.Glob("*.hurl")
	if err != nil {
		t.Fatalf("Failed to list .hurl files: %v", err)
	}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			cmd := exec.Command("hurl", file)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("hurl command failed for %s: %v\nOutput: %s", file, err, output)
			}
		})
	}

}

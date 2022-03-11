package inference

import (
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestEvaluateInferenceRule(t *testing.T) {
	val, err := EvaluateInferenceRule(`[{"steps": [{"root": "/"}]}]`)
	if err != nil {
		t.Fatal(err)
	}
	if val != nil {
		t.Errorf("got v=%+#v", val)
	}
}

func TestEvaluateInferenceRule_2(t *testing.T) {
	ts := []struct {
		rule       string
		wantOutput []config.IndexJob
	}{
		{rule: `
		var blockedSegments = ["example", "examples", "vendor"]; // TODO add more
		function dirname(path) {
			var fileSeparator = path.lastIndexOf("/")
			if (fileSeparator === -1) {
				return "/"
			}
			return path.substring(0, fileSeparator)
		}
		function basename(path) {
			var fileSeparator = path.lastIndexOf("/")
			if (fileSeparator === -1) {
				return path
			}
			return path.substring(fileSeparator + 1)
		}
		function isGoModulePath(path) {
			return basename(path) === "go.mod" && !blockedSegments.some(function(segment) { return path.split("/").some(function(dir) { return dir === segment }) })
		};

		lsFiles("go.mod").filter(isGoModulePath).map(function(path) {
			return {
				"steps": [{
					"root": dirname(path),
					"image": "lsif-go:whatever",
					"commands": ["go mod download"]
				}],
				"root": dirname(path)
			}
		});
		`,
			wantOutput: []config.IndexJob{{
				Steps: []config.DockerStep{
					{
						Root:     "/erick",
						Image:    "lsif-go:whatever",
						Commands: []string{"go mod download"},
					},
				},
				Root:        "/erick",
				LocalSteps:  []string{},
				IndexerArgs: []string{},
			}},
		},
		// {rule: `
		// {rule: ``},
		// {rule: ``},
		// {rule: ``},
		// {rule: ``},
	}

	for idx, tt := range ts {
		t.Run(strconv.Itoa(idx), func(t *testing.T) {
			have, err := EvaluateInferenceRule(tt.rule)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.wantOutput, have); diff != "" {
				t.Fatalf("invalid output received: %s", diff)
			}
		})
	}
}

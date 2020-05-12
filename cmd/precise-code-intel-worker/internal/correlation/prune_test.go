package correlation

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/datastructures"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/lsif"
)

func TestPrune(t *testing.T) {
	gitContentsOracle := map[string][]string{
		"root":     {"root/sub/", "root/foo.go", "root/bar.go"},
		"root/sub": {"root/sub/baz.go"},
	}

	getChildren := func(dirnames []string) (map[string][]string, error) {
		out := map[string][]string{}
		for _, dirname := range dirnames {
			out[dirname] = gitContentsOracle[dirname]
		}

		return out, nil
	}

	state := &State{
		DocumentData: map[string]lsif.DocumentData{
			"d01": {URI: "foo.go"},
			"d02": {URI: "bar.go"},
			"d03": {URI: "sub/baz.go"},
			"d04": {URI: "foo.generated.go"},
			"d05": {URI: "foo.generated.go"},
		},
		DefinitionData: map[string]datastructures.DefaultIDSetMap{
			"x01": {"d01": {}, "d04": {}},
			"x02": {"d02": {}},
		},
		ReferenceData: map[string]datastructures.DefaultIDSetMap{
			"x03": {"d02": {}},
			"x04": {"d02": {}, "d05": {}},
		},
	}

	if err := prune(state, "root", getChildren); err != nil {
		t.Fatalf("unexpected error pruning state: %s", err)
	}

	expectedState := &State{
		DocumentData: map[string]lsif.DocumentData{
			"d01": {URI: "foo.go"},
			"d02": {URI: "bar.go"},
			"d03": {URI: "sub/baz.go"},
		},
		DefinitionData: map[string]datastructures.DefaultIDSetMap{
			"x01": {"d01": {}},
			"x02": {"d02": {}},
		},
		ReferenceData: map[string]datastructures.DefaultIDSetMap{
			"x03": {"d02": {}},
			"x04": {"d02": {}},
		},
	}
	if diff := cmp.Diff(expectedState, state); diff != "" {
		t.Errorf("unexpected state (-want +got):\n%s", diff)
	}
}

package correlation

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/correlation/datastructures"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/correlation/lsif"
)

func TestPrune(t *testing.T) {
	gitContentsOracle := map[string][]string{
		"root":     {"root/sub/", "root/foo.go", "root/bar.go"},
		"root/sub": {"root/sub/baz.go"},
	}

	getChildren := func(ctx context.Context, dirnames []string) (map[string][]string, error) {
		out := map[string][]string{}
		for _, dirname := range dirnames {
			out[dirname] = gitContentsOracle[dirname]
		}

		return out, nil
	}

	state := &State{
		DocumentData: map[int]lsif.Document{
			1001: {
				URI:         "foo.go",
				Contains:    datastructures.NewIDSet(),
				Diagnostics: datastructures.NewIDSet(),
			},
			1002: {
				URI:         "bar.go",
				Contains:    datastructures.NewIDSet(),
				Diagnostics: datastructures.NewIDSet(),
			},
			1003: {
				URI:         "sub/baz.go",
				Contains:    datastructures.NewIDSet(),
				Diagnostics: datastructures.NewIDSet(),
			},
			1004: {
				URI:         "foo.generated.go",
				Contains:    datastructures.NewIDSet(),
				Diagnostics: datastructures.NewIDSet(),
			},
			1005: {
				URI:         "foo.generated.go",
				Contains:    datastructures.NewIDSet(),
				Diagnostics: datastructures.NewIDSet(),
			},
		},
		DefinitionData: map[int]datastructures.DefaultIDSetMap{
			2001: {
				1001: datastructures.NewIDSet(),
				1004: datastructures.NewIDSet(),
			},
			2002: {
				1002: datastructures.NewIDSet(),
			},
		},
		ReferenceData: map[int]datastructures.DefaultIDSetMap{
			2003: {
				1002: datastructures.NewIDSet(),
			},
			2004: {
				1002: datastructures.NewIDSet(),
				1005: datastructures.NewIDSet(),
			},
		},
	}

	if err := prune(context.Background(), state, "root", getChildren); err != nil {
		t.Fatalf("unexpected error pruning state: %s", err)
	}

	expectedState := &State{
		DocumentData: map[int]lsif.Document{
			1001: {
				URI:         "foo.go",
				Contains:    datastructures.NewIDSet(),
				Diagnostics: datastructures.NewIDSet(),
			},
			1002: {
				URI:         "bar.go",
				Contains:    datastructures.NewIDSet(),
				Diagnostics: datastructures.NewIDSet(),
			},
			1003: {
				URI:         "sub/baz.go",
				Contains:    datastructures.NewIDSet(),
				Diagnostics: datastructures.NewIDSet(),
			},
		},
		DefinitionData: map[int]datastructures.DefaultIDSetMap{
			2001: {
				1001: datastructures.NewIDSet(),
			},
			2002: {
				1002: datastructures.NewIDSet(),
			},
		},
		ReferenceData: map[int]datastructures.DefaultIDSetMap{
			2003: {
				1002: datastructures.NewIDSet(),
			},
			2004: {
				1002: datastructures.NewIDSet(),
			},
		},
	}
	if diff := cmp.Diff(expectedState, state, datastructures.IDSetComparer); diff != "" {
		t.Errorf("unexpected state (-want +got):\n%s", diff)
	}
}

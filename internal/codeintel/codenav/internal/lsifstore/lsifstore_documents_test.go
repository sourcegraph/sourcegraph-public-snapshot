package lsifstore

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/symbols"
)

func TestGetFullSCIPNameByDescriptor(t *testing.T) {
	ctx := context.Background()
	store := populateTestStore(t)
	uploadIDs := []int{testSCIPUploadID2}
	symbolNames := []string{
		"treesitter . . . jen/Statement#Add().",
		"treesitter . . . yamlPayload#Force.",
		"treesitter . . . internal/mockgen/generation/generateMockStructFromConstructorCommon().",
		"treesitter . . . testutil/require/CalledWith().",
		"treesitter . . . nomatch.",
		"malformed symbol",
	}

	explodedSymbols, err := store.GetFullSCIPNameByDescriptor(ctx, uploadIDs, symbolNames)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	expected := []*symbols.ExplodedSymbol{
		{
			Scheme:         "scip-go",
			PackageManager: "gomod",
			PackageName:    "github.com/dave/jennifer",
			PackageVersion: "v1.5.0",
			Descriptor:     "github.com/dave/jennifer/jen/Statement#Add().",
		},
		{
			Scheme:         "scip-go",
			PackageManager: "gomod",
			PackageName:    "github.com/derision-test/go-mockgen",
			PackageVersion: "d061eb01e698",
			Descriptor:     "github.com/derision-test/go-mockgen/cmd/go-mockgen/yamlPayload#Force.",
		},
		{
			Scheme:         "scip-go",
			PackageManager: "gomod",
			PackageName:    "github.com/derision-test/go-mockgen",
			PackageVersion: "d061eb01e698",
			Descriptor:     "github.com/derision-test/go-mockgen/internal/mockgen/generation/generateMockStructFromConstructorCommon().",
		},
		{
			Scheme:         "scip-go",
			PackageManager: "gomod",
			PackageName:    "github.com/derision-test/go-mockgen",
			PackageVersion: "d061eb01e698",
			Descriptor:     "github.com/derision-test/go-mockgen/testutil/require/CalledWith().",
		},

		//
		// Over-selection:

		{
			Scheme:         "scip-go",
			PackageManager: "gomod",
			PackageName:    "github.com/derision-test/go-mockgen",
			PackageVersion: "d061eb01e698",
			Descriptor:     "github.com/derision-test/go-mockgen/testutil/require/NotCalledWith().",
		},
		{
			Scheme:         "scip-go",
			PackageManager: "gomod",
			PackageName:    "github.com/derision-test/go-mockgen",
			PackageVersion: "d061eb01e698",
			Descriptor:     "github.com/derision-test/go-mockgen/testutil/assert/CalledWith().",
		},
		{
			Scheme:         "scip-go",
			PackageManager: "gomod",
			PackageName:    "github.com/derision-test/go-mockgen",
			PackageVersion: "d061eb01e698",
			Descriptor:     "github.com/derision-test/go-mockgen/testutil/assert/NotCalledWith().",
		},
		{
			Scheme:         "scip-go",
			PackageManager: "gomod",
			PackageName:    "github.com/derision-test/go-mockgen",
			PackageVersion: "d061eb01e698",
			Descriptor:     "github.com/derision-test/go-mockgen/testutil/gomega/BeCalledWith().",
		},
	}

	sort.Slice(expected, func(i, j int) bool { return expected[i].Descriptor < expected[j].Descriptor })
	sort.Slice(explodedSymbols, func(i, j int) bool { return explodedSymbols[i].Descriptor < explodedSymbols[j].Descriptor })

	if diff := cmp.Diff(expected, explodedSymbols); diff != "" {
		t.Errorf("unexpected symbols (-want +got):\n%s", diff)
	}
}

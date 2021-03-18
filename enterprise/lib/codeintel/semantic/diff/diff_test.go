package diff

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/hexops/autogold"
	"github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/lsif/conversion"
	"github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/semantic"
)

var dumpPath = "./testdata/project1/dump.lsif"
var dumpPermutedPath = "./testdata/project1/dump-permuted.lsif"
var dumpOldPath = "./testdata/project1/dump-old.lsif"
var dumpNewPath = "./testdata/project1/dump-new.lsif"

func TestNoDiffOnPermutedDumps(t *testing.T) {
	bundle1, err := conversion.CorrelateLocalGit(
		context.Background(),
		dumpPath,
		filepath.Dir(dumpPath),
	)
	if err != nil {
		t.Fatalf("Unexpected error reading dump path: %v", err)
	}

	bundle2, err := conversion.CorrelateLocalGit(
		context.Background(),
		dumpPermutedPath,
		filepath.Dir(dumpPermutedPath),
	)
	if err != nil {
		t.Fatalf("Unexpected error reading dump path: %v", err)
	}

	diff := Diff(semantic.GroupedBundleDataChansToMaps(bundle1), semantic.GroupedBundleDataChansToMaps(bundle2))

	if diff != "" {
		t.Fatalf("Expected semantic.Diff to compute that dumps %v and %v are semantically equal, got:\n%v", dumpPath, dumpPermutedPath, diff)
	}
}

func TestDiffOnEditedDumps(t *testing.T) {
	bundle1, err := conversion.CorrelateLocalGit(
		context.Background(),
		dumpOldPath,
		filepath.Dir(dumpOldPath),
	)
	if err != nil {
		t.Fatalf("Unexpected error reading dump: %v", err)
	}

	bundle2, err := conversion.CorrelateLocalGit(
		context.Background(),
		dumpNewPath,
		filepath.Dir(dumpNewPath),
	)
	if err != nil {
		t.Fatalf("Unexpected error reading dump: %v", err)
	}

	computedDiff := Diff(
		semantic.GroupedBundleDataChansToMaps(bundle1),
		semantic.GroupedBundleDataChansToMaps(bundle2),
	)

	autogold.Equal(t, autogold.Raw(computedDiff))
}

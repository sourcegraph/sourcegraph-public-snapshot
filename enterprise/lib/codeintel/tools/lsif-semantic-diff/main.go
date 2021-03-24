package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/lsif/conversion"
	"github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/semantic"
	"github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/semantic/diff"
)

func main() {
	if len(os.Args) != 3 {
		usage()
	}

	dumpPath1 := os.Args[1]
	dumpPath2 := os.Args[2]

	if !strings.HasSuffix(dumpPath1, ".lsif") {
		usage()
	}
	if !strings.HasSuffix(dumpPath2, ".lsif") {
		usage()
	}

	bundle1, err := conversion.CorrelateLocalGit(
		context.Background(),
		dumpPath1,
		filepath.Dir(dumpPath1),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	bundle2, err := conversion.CorrelateLocalGit(
		context.Background(),
		dumpPath2,
		filepath.Dir(dumpPath2),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(diff.Diff(
		semantic.GroupedBundleDataChansToMaps(bundle1),
		semantic.GroupedBundleDataChansToMaps(bundle2),
	))
}

func usage() {
	fmt.Println(`
usage: diff old.lsif new.lsif

lsif dumps must be in project's root directory`)
	os.Exit(1)
}

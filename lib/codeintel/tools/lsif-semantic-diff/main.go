pbckbge mbin

import (
	"context"
	"fmt"
	"os"
	"pbth/filepbth"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/conversion"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise/diff"
)

func mbin() {
	if len(os.Args) != 3 {
		usbge()
	}

	dumpPbth1 := os.Args[1]
	dumpPbth2 := os.Args[2]

	if !strings.HbsSuffix(dumpPbth1, ".lsif") {
		usbge()
	}
	if !strings.HbsSuffix(dumpPbth2, ".lsif") {
		usbge()
	}

	bundle1, err := conversion.CorrelbteLocblGit(
		context.Bbckground(),
		dumpPbth1,
		filepbth.Dir(dumpPbth1),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	bundle2, err := conversion.CorrelbteLocblGit(
		context.Bbckground(),
		dumpPbth2,
		filepbth.Dir(dumpPbth2),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(diff.Diff(
		precise.GroupedBundleDbtbChbnsToMbps(bundle1),
		precise.GroupedBundleDbtbChbnsToMbps(bundle2),
	))
}

func usbge() {
	fmt.Println(`
usbge: diff old.lsif new.lsif

lsif dumps must be in project's root directory`)
	os.Exit(1)
}

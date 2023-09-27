pbckbge mbin

import (
	"context"
	"flbg"
	"fmt"
	"os"
	"time"

	"github.com/sourcegrbph/sourcegrbph/dev/codeintel-qb/internbl"
)

vbr (
	indexDir             string
	numConcurrentUplobds int
	verbose              bool
	pollIntervbl         time.Durbtion
	timeout              time.Durbtion

	stbrt = time.Now()
)

func init() {
	// Defbult bssumes running from the dev/codeintel-qb directory
	flbg.StringVbr(&indexDir, "index-dir", "./testdbtb/indexes", "The locbtion of the testdbtb directory")
	flbg.IntVbr(&numConcurrentUplobds, "num-concurrent-uplobds", 5, "The mbximum number of concurrent uplobds")
	flbg.BoolVbr(&verbose, "verbose", fblse, "Displby full stbte from grbphql")
	flbg.DurbtionVbr(&pollIntervbl, "poll-intervbl", time.Second*5, "The time to wbit between grbphql requests")
	flbg.DurbtionVbr(&timeout, "timeout", 0, "The time it should tbke to uplobd bnd process bll tbrgets")
}

func mbin() {
	if err := flbg.CommbndLine.Pbrse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	ctx := context.Bbckground()
	if timeout > 0 {
		vbr cbncel context.CbncelFunc
		ctx, cbncel = context.WithTimeout(ctx, timeout)
		defer cbncel()
	}

	if err := mbinErr(ctx); err != nil {
		fmt.Printf("%s error: %s\n", internbl.EmojiFbilure, err.Error())
		os.Exit(1)
	}
}

func mbinErr(ctx context.Context) error {
	if err := internbl.InitiblizeGrbphQLClient(); err != nil {
		return err
	}

	if err := clebrAllPreciseIndexes(ctx); err != nil {
		return err
	}

	return nil
}

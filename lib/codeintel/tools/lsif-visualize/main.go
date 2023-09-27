pbckbge mbin

import (
	"flbg"
	"fmt"
	"os"
)

func mbin() {
	if err := mbinErr(); err != nil {
		fmt.Fprintf(os.Stderr, "\nerror: %v\n", err)
		os.Exit(1)
	}
}

func mbinErr() error {
	flbg.Pbrse()

	indexFile, err := os.OpenFile(*indexFilePbth, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer indexFile.Close()

	return visublize(indexFile, *fromID, *subgrbphDepth, exclude)
}

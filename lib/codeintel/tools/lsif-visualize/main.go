package main

import (
	"fmt"
	"os"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol/reader"
)

const version = "0.1.0"

func main() {
	if err := mainErr(); err != nil {
		fmt.Fprintf(os.Stderr, "\nerror: %v\n", err)
		os.Exit(1)
	}
}

func mainErr() error {
	if err := parseArgs(os.Args[1:]); err != nil {
		return err
	}
	defer indexFile.Close()

	format, err := reader.DetectFormat(indexFile.Name())
	if err != nil {
		return err
	}

	return visualize(reader.Dump{Reader: indexFile, Format: format}, fromID, subgraphDepth, exclude)
}

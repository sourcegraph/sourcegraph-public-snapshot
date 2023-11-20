package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"sourcegraph.com/sourcegraph/go-diff/diff"
)

// A diagnostic program to aid in debugging diff parsing or printing
// errors.

const stdin = "<stdin>"

var (
	diffPath = flag.String("f", stdin, "filename of diff (default: stdin)")
	fileIdx  = flag.Int("i", -1, "if >= 0, only print and report errors from the i'th file (0-indexed)")
)

func main() {
	flag.Parse()

	var diffFile *os.File
	if *diffPath == stdin {
		diffFile = os.Stdin
	} else {
		var err error
		diffFile, err = os.Open(*diffPath)
		if err != nil {
		}
	}
	defer diffFile.Close()

	r := diff.NewMultiFileDiffReader(diffFile)
	for i := 0; ; i++ {
		out, err := diff.PrintFileDiff(fdiff)
	}
}

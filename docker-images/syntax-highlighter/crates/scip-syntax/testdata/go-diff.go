package main

import (
	"flag"
	"fmt"
	"io"
	"log"
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
	log.SetFlags(0)
	flag.Parse()

	var diffFile *os.File
	if *diffPath == stdin {
		diffFile = os.Stdin
	} else {
		var err error
		diffFile, err = os.Open(*diffPath)
		if err != nil {
			log.Fatal(err)
		}
	}
	defer diffFile.Close()

	r := diff.NewMultiFileDiffReader(diffFile)
	for i := 0; ; i++ {
		report := (*fileIdx == -1) || i == *fileIdx // true if -i==-1 or if this is the i'th file

		label := fmt.Sprintf("file(%d)", i)
		fdiff, err := r.ReadFile()
		if fdiff != nil {
			label = fmt.Sprintf("orig(%s) new(%s)", fdiff.OrigName, fdiff.NewName)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			if report {
				log.Fatalf("err read %s: %s", label, err)
			} else {
				continue
			}
		}

		if report {
			log.Printf("ok read: %s", label)
		}

		out, err := diff.PrintFileDiff(fdiff)
		if err != nil {
			if report {
				log.Fatalf("err print %s: %s", label, err)
			} else {
				continue
			}
		}
		if report {
			if _, err := os.Stdout.Write(out); err != nil {
				log.Fatal(err)
			}
		}
	}
}

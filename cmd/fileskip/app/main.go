package main

import (
	"bytes"
	"fmt"
	"github.com/bits-and-blooms/bloom/v3"
	"github.com/cockroachdb/errors"
	"github.com/sourcegraph/sourcegraph/cmd/fileskip"
	"os"
)

var cacheFile = os.Getenv("HOME") + "/dev/sourcegraph/fileskip-cache"

func main2() {
	b := bloom.NewWithEstimates(100_000, 0.01)
	var buf bytes.Buffer
	_, err := b.WriteTo(&buf)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(os.Getenv("HOME")+"/dev/sourcegraph/cache", buf.Bytes(), 0755)
	if err != nil {
		panic(err)
	}
}

func main() {
	fmt.Println(os.Args)
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "index":
			dir := os.Args[2]
			err := fileskip.WriteCache(dir, cacheFile)
			if err != nil {
				panic(err)
			}
		case "grep":
			r, err := fileskip.ReadCache(cacheFile)
			if err != nil {
				panic(errors.Wrapf(err, "Failed to decode cache at %v", cacheFile))
			}
			for _, arg := range os.Args[2:] {
				r.Grep(arg)
			}
		default:
			fmt.Printf("unknown command '%v'. Expected 'index' or 'grep'", os.Args[1])
			os.Exit(1)
		}
	}
	//QueryCache()
}

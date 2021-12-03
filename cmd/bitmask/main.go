package main

import (
	"bytes"
	"fmt"
	"github.com/bits-and-blooms/bloom/v3"
	"os"
)

var cacheDir = os.Getenv("HOME") + "/dev/sourcegraph/bitmask-cache"

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
			err := WriteCache(dir, cacheDir)
			if err != nil {
				panic(err)
			}
		case "grep":
			r := ReadCache(cacheDir)
			for _, arg := range os.Args[2:] {
				fmt.Printf("query=%v\n", arg)
				r.Grep(arg)
			}
		}
	}
	//QueryCache()
}

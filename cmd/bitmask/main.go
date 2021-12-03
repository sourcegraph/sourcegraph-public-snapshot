package main

import (
	"fmt"
	"os"
)

func main() {
	var dir string
	//dir := os.Getenv("HOME") + "/dev/sgtest/megarepo"
	fmt.Println(os.Args)
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "index":
			dir = os.Getenv("HOME") + "/dev/sourcegraph/sourcegraph"
			WriteCache(dir, 1_000_000)
		case "grep":
			r := ReadCache()
			for _, arg := range os.Args[2:] {
				fmt.Printf("query=%v\n", arg)
				r.Grep(arg)
			}
		}
	}
	//QueryCache()
}

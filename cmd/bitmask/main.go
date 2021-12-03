package main

import "os"

func main() {
	var dir string
	//dir := os.Getenv("HOME") + "/dev/sgtest/megarepo"
	dir = os.Getenv("HOME") + "/dev/sourcegraph/sourcegraph"
	WriteCache(dir, 1_000_000)
	//QueryCache()
}

package main

import (
	"flag"
)

var (
	indexFilePath = flag.String("index-file", "dump.lsif", "The LSIF index to validate.")
)

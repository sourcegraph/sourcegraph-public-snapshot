package main

import (
	"flag"
)

var (
	indexFilePath  = flag.String("index-file", "dump.lsif", "The LSIF index to rank.")
	outputFilePath = flag.String("output-file", "", "Output file; defaults to input + '-pagerank'.")
	addImplEdges   = flag.Bool("include-impls", false, "True to include implementation edges")
)

package main

import (
	"flag"
	"strings"
)

var (
	indexFilePath = flag.String("index-file", "dump.lsif", "The LSIF index to visualize.")
	fromID        = flag.Int("from-id", 2, "The edge/vertex ID to visualize a subgraph from. Must be used in combination with '-depth'.")
	subgraphDepth = flag.Int("depth", -1, "Depth limit of the subgraph to be output")
	excludeArg    = flag.String("exclude", "", "Comma-separated list of vertices to exclude from the visualization")
	exclude       = strings.Split(*excludeArg, ",")
)

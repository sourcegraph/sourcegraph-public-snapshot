package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
)

// This tool reads the specified lsif file and prints pagerank-sorted
// file paths in space-delimited output to stdout:  filepath rank
func main() {
	flag.Parse()

	indexFile, err := os.OpenFile(*indexFilePath, os.O_RDONLY, 0)
	if err != nil {
		die("Unable to open index file %v: %v", *indexFilePath, err)
	}
	defer indexFile.Close()

	rankings, err := PageRankLsif(indexFile)
	if err != nil {
		die("Error computing pagerank: %v", err)
	}

	// For now, write file names and their rankings (sorted) to stdout.
	sorter := *rankings
	sort.Slice(sorter, func(i, j int) bool {
		return sorter[i].rank > sorter[j].rank
	})

	for _, doc := range sorter {
		fmt.Printf("%v  %v\n", doc.filePath, doc.rank)
	}
}

func die(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, "\nerror: "+fmt.Sprintf(msg, args))
	os.Exit(1)
}

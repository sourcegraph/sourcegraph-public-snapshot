package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if err := mainErr(); err != nil {
		fmt.Fprintf(os.Stderr, "\nerror: %v\n", err)
		os.Exit(1)
	}
}

func mainErr() error {
	flag.Parse()

	indexFile, err := os.OpenFile(*indexFilePath, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer indexFile.Close()

	return validate(indexFile)
}

package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

func main() {
	if err := mainErr(); err != nil {
		fmt.Fprint(os.Stderr, fmt.Sprintf("\nerror: %v\n", err))
		os.Exit(1)
	}
}

func mainErr() error {
	if err := parseArgs(os.Args[1:]); err != nil {
		return err
	}
	defer indexFile.Close()

	return validate(indexFile)
}

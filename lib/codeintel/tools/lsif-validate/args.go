package main

import (
	"os"

	"github.com/alecthomas/kingpin"
)

var app = kingpin.New(
	"lsif-validate",
	"lsif-validate is validator for LSIF indexer output.",
).Version(version)

var (
	indexFile *os.File
)

func init() {
	app.HelpFlag.Short('h')
	app.VersionFlag.Short('v')
	app.HelpFlag.Hidden()

	app.Arg("index-file", "The LSIF index to validate.").Default("dump.lsif").FileVar(&indexFile)
}

func parseArgs(args []string) (err error) {
	if _, err := app.Parse(args); err != nil {
		return err
	}

	return nil
}

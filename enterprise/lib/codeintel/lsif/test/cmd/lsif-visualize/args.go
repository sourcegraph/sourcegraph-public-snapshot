package main

import (
	"os"

	"github.com/alecthomas/kingpin"
)

var app = kingpin.New(
	"lsif-visualize",
	"lsif-visualize is visualizer for LSIF indexer output.",
).Version(version)

var (
	indexFile     *os.File
	fromID        int
	subgraphDepth int
)

func init() {
	app.HelpFlag.Short('h')
	app.VersionFlag.Short('v')
	app.HelpFlag.Hidden()

	app.Flag("from-id", "The edge/vertex ID to visualize a subgraph from. Must be used in combination with '-depth'.").Default("2").IntVar(&fromID)
	app.Flag("depth", "Depth limit of the subgraph to be output").Default("-1").IntVar(&subgraphDepth)

	app.Arg("index-file", "The LSIF index to visualize.").Default("dump.lsif").FileVar(&indexFile)
}

func parseArgs(args []string) (err error) {
	if _, err := app.Parse(args); err != nil {
		return err
	}

	return nil
}

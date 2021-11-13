package main

import (
	"os"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol/reader"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/tools/lsif-visualize/internal/visualization"
)

func visualize(dump reader.Dump, fromID, subgraphDepth int, exclude []string) error {
	ctx := visualization.NewVisualizationContext()
	visualizer := &visualization.Visualizer{Context: ctx}
	return visualizer.Visualize(dump, fromID, subgraphDepth, exclude)
}

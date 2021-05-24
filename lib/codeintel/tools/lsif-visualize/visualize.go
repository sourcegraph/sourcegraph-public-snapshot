package main

import (
	"os"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/tools/lsif-visualize/internal/visualization"
)

func visualize(indexFile *os.File, fromID, subgraphDepth int, exclude []string) error {
	ctx := visualization.NewVisualizationContext()
	visualizer := &visualization.Visualizer{Context: ctx}
	return visualizer.Visualize(indexFile, fromID, subgraphDepth, exclude)
}

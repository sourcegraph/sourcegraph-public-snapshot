package main

import (
	"os"

	"github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/lsif/test/cmd/lsif-visualize/internal/visualization"
)

func visualize(indexFile *os.File, fromID, subgraphDepth int) error {
	ctx := visualization.NewVisualizationContext()
	visualizer := &visualization.Visualizer{Context: ctx}
	return visualizer.Visualize(indexFile, fromID, subgraphDepth)
}

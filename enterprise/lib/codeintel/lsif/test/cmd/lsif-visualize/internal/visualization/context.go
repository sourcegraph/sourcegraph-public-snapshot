package visualization

import (
	"github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/lsif/test/internal/reader"
)

type VisualizationContext struct {
	Stasher *reader.Stasher
}

func NewVisualizationContext() *VisualizationContext {
	return &VisualizationContext{
		Stasher: reader.NewStasher(),
	}
}

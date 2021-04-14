package visualization

import (
	"github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/lsif/internal/reader"
)

type VisualizationContext struct {
	Stasher *reader.Stasher
}

func NewVisualizationContext() *VisualizationContext {
	return &VisualizationContext{
		Stasher: reader.NewStasher(),
	}
}

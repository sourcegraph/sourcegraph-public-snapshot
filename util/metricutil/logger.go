package metricutil

import (
	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
)

type logger struct {
}

func (l *logger) Log(ctx context.Context, event *sourcegraph.UserEvent) {
	// currently discarded
}

func LogEvent(ctx context.Context, event *sourcegraph.UserEvent) {
	// currently discarded
}

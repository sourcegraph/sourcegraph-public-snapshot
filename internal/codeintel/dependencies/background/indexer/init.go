package indexer

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewIndexer() goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.Interval, &indexer{})
}

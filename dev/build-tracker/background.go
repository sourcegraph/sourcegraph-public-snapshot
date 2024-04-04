package main

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/dev/build-tracker/build"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func deleteOldBuilds(logger log.Logger, store *build.Store, every, window time.Duration) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), goroutine.HandlerFunc(func(ctx context.Context) error {
		oldBuilds := make([]int, 0)
		now := time.Now()
		for _, b := range store.FinishedBuilds() {
			finishedAt := *b.FinishedAt
			delta := now.Sub(finishedAt.Time)
			if delta >= window {
				logger.Debug("build past age window", log.Int("buildNumber", *b.Number), log.Time("FinishedAt", finishedAt.Time), log.Duration("window", window))
				oldBuilds = append(oldBuilds, *b.Number)
			}
		}
		logger.Info("deleting old builds", log.Int("oldBuildCount", len(oldBuilds)))
		store.DelByBuildNumber(oldBuilds...)
		return nil
	}), goroutine.WithInterval(every), goroutine.WithName("old-build-purger"))
}

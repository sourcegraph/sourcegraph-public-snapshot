package analytics

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/background"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

var (
	bq *BigQueryClient
)

func BackgroundEventPublisher(ctx context.Context) {
	background.Run(ctx, func(ctx context.Context, bgOut *std.Output) {
		var err error
		bq, err = NewBigQueryClient(ctx, SGLocalDev, AnalyticsDatasetName, EventsTableName)
		if err != nil {
			bgOut.WriteWarningf("failed to create BigQuery client for analytics", err)
			return
		}

		var store analyticsStore
		done := make(chan struct{})
		go processEvents(bgOut, store, done)
	})
}

func toEvents(items []invocation) []event {
	results := []event{}
	for _, i := range items {
		ev := NewEvent(i)
		results = append(results, *ev)
	}

	return results
}

func processEvents(bgOut *std.Output, store analyticsStore, done chan struct{}) error {
	ctx := context.Background()

	tickRate := 10 * time.Millisecond
	ticker := time.NewTicker(tickRate)

	for {
		select {
		case <-ticker.C:
			ticker.Stop()
			results, err := store.ListCompleted(ctx)
			if err != nil {
				bgOut.WriteWarningf("failed to list completed analytics events", err)
				continue
			}

			events := toEvents(results)
			for _, ev := range events {
				err := bq.InsertEvent(ctx, ev)
				if err != nil {
					bgOut.WriteWarningf("failed to insert analytics event", err)
				}
				store.DeleteInvocation(ctx, ev.UUID)
			}

			ticker.Reset(tickRate)
		case <-done:
			return nil
		}
	}

}

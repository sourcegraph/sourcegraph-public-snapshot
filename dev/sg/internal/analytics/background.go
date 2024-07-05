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

func BackgroundEventPublisher(ctx context.Context) func() {
	done := make(chan struct{})
	background.Run(ctx, func(ctx context.Context, bgOut *std.Output) {
		println("💀 BG START")
		var err error
		bq, err = NewBigQueryClient(ctx, SGLocalDev, AnalyticsDatasetName, EventsTableName)
		if err != nil {
			bgOut.WriteWarningf("failed to create BigQuery client for analytics", err)
			return
		}

		go processEvents(bgOut, store(), done)
	})

	return func() {
		close(done)
	}
}

func toEvents(items []invocation) []event {
	results := []event{}
	for _, i := range items {
		ev := NewEvent(i)
		results = append(results, *ev)
	}

	return results
}

func processEvents(bgOut *std.Output, store analyticsStore, done chan struct{}) {
	println("💀 PROCESS EVENTS")
	ctx := context.Background()

	for {
		select {
		case <-done:
			return
		default:
			results, err := store.ListCompleted(ctx)
			if err != nil {
				bgOut.WriteWarningf("failed to list completed analytics events", err)
				// TODO(burmudar): We sleep here for now, but we need to try about
				// 3 times and stop and print out that we stopped because there is something big wrong
				time.Sleep(time.Second)
				continue
			}

			events := toEvents(results)
			println("💀 EVENTS", len(events))
			for _, ev := range events {
				err := bq.InsertEvent(ctx, ev)
				if err != nil {
					println("💀💀💀💀", err)
					panic(err)
					bgOut.WriteWarningf("failed to insert analytics event", err)
					continue
				}
				store.DeleteInvocation(ctx, ev.UUID)
			}
		}
	}

}

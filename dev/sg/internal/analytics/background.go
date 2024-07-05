package analytics

import (
	"context"
	"os"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/background"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

var (
	bq   *BigQueryClient
	done chan struct{}
)

func BackgroundEventPublisher(ctx context.Context) {
	done := make(chan struct{})
	background.Run(ctx, func(ctx context.Context, bgOut *std.Output) {
		var err error
		bq, err = NewBigQueryClient(ctx, SGLocalDev, AnalyticsDatasetName, EventsTableName)
		if err != nil {
			bgOut.WriteWarningf("failed to create BigQuery client for analytics", err)
			return
		}
		defer bq.Close()

		processEvents(bgOut, store(), done)
	})
}

func StopBackgroundEventPublisher() {
	close(done)
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

			if len(results) == 0 {
				// No events to process - so we quit.
				//
				// Upon next start up there will be another event to publish
				return
			}

			events := toEvents(results)
			for _, ev := range events {
				err := bq.InsertEvent(ctx, ev)
				if err != nil {
					if os.Getenv("SG_ANALYTICS_DEBUG") == "1" {
						panic(err)
					}
					bgOut.WriteWarningf("failed to insert analytics event", err)
					continue
				}
				store.DeleteInvocation(ctx, ev.UUID)
			}
			// all events have been processed, so we quit
			return
		}
	}

}

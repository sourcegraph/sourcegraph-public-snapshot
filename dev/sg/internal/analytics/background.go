package analytics

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/background"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	bq   *BigQueryClient
	done chan struct{}
)

func BackgroundEventPublisher(ctx context.Context) {
	done = make(chan struct{})
	background.Run(ctx, func(ctx context.Context, bgOut *std.Output) {
		var err error
		bq, err = NewBigQueryClient(ctx, SGLocalDev, AnalyticsDatasetName, EventsTableName)
		if err != nil {
			bgOut.WriteWarningf("failed to create BigQuery client for analytics", err)
			return
		}
		defer bq.Close()

		processEvents(ctx, bgOut, store(), done)
	})
}

func StopBackgroundEventPublisher() {
	close(done)
}

func toEvents(items []invocation) []event {
	results := make([]event, 0, len(items))
	for _, i := range items {
		ev := NewEvent(i)
		results = append(results, *ev)
	}

	return results
}

func maybePrintHelpMsg(out *std.Output, multErr errors.MultiError) {
	if multErr == nil {
		return
	}
	errs := multErr.Errors()
	if len(errs) == 0 {
		return
	}
	errMsg := ""
	for _, e := range errs {
		errMsg += e.Error()
		errMsg += "\n"
	}

	out.WriteWarningf("%d Errors occured while trying to publish analytics to bigquery. Below are some of the errors:", len(errs))
	msg := fmt.Sprintf("\n```%s```\n", errMsg)
	msg += "If these errors persist you can disable analytics with `export SG_DISABLE_ANALYTICS=1` or by passing the flag `--disable-analytics` as part of your command\n"
	msg += "Alternatively, try one of the following:"
	msg += "- You should be in the `gcp-engineering@sourcegraph.com` group. Ask #ask-it-tech-ops or #discuss-dev-infra to check that\n"
	msg += "- Ensure you're currently authenticated with your sourcegraph.com account by running `gcloud auth list`\n"
	msg += "- Ensure you're authenticated with gcloud by running `gcloud auth application-default login`\n"
	out.WriteMarkdown(msg)
}

func processEvents(ctx context.Context, bgOut *std.Output, store analyticsStore, done chan struct{}) {
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
			var errs errors.MultiError
			for _, ev := range events {
				err := bq.InsertEvent(ctx, ev)
				if err != nil {
					if os.Getenv("SG_ANALYTICS_DEBUG") == "1" {
						panic(err)
					}
					errs = errors.Append(errs, err)
					continue
				}

				err = store.DeleteInvocation(ctx, ev.UUID)
				if err != nil {
					errs = errors.Append(errs, err)
				}

				if errs != nil && len(errs.Errors()) > 3 {
					// if we have more than 3 errors. Something is wrong and it's better for us to exit early.
					break
				}
			}

			maybePrintHelpMsg(bgOut, errs)
		}
	}

}

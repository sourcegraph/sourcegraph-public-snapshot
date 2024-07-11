package server

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/telemetry-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/lib/telemetrygateway/v1"
)

func TestSummarizeFailedEvents(t *testing.T) {
	t.Run("all failed", func(t *testing.T) {
		const failureMessage = "event publish failed"
		submitted := make([]*telemetrygatewayv1.Event, 5)
		for i := range submitted {
			submitted[i] = &telemetrygatewayv1.Event{
				Id:      "id_" + strconv.Itoa(i),
				Feature: "feature_" + strconv.Itoa(i),
				Action:  "action_" + strconv.Itoa(i),
				Source: &telemetrygatewayv1.EventSource{
					Server: &telemetrygatewayv1.EventSource_Server{Version: t.Name()},
				},
			}
			if i%2 == 0 {
				submitted[i].Source.Client = &telemetrygatewayv1.EventSource_Client{
					Name: "test_client",
				}
			}
		}
		results := make([]events.PublishEventResult, len(submitted))
		for i, event := range submitted {
			results[i] = events.NewPublishEventResult(event, errors.New(failureMessage), true)
		}
		require.False(t, len(results) == 0)

		summary := summarizePublishEventsResults(results, summarizePublishEventsResultsOpts{})
		autogold.Expect("all events in batch failed to submit").Equal(t, summary.message)
		autogold.Expect("complete_failure").Equal(t, summary.result)
		assert.Len(t, summary.errorFields, len(results))
		assert.Len(t, summary.succeededEvents, 0)
		assert.Len(t, summary.failedEvents, len(results))

		// This sub-test asserts some behaviour specific to how we construct
		// Sentry reports in sourcegraph/log:
		// https://github.com/sourcegraph/log/blob/f53023898988779b0dabb75fda2c5c3b8d5ae3ae/internal/sinkcores/sentrycore/worker.go#L95-L96
		//
		// It requires knowledge of the internals of the logging library and
		// how Sentry interprets these reports.
		// TODO: https://github.com/sourcegraph/log/issues/65
		t.Run("Sentry report", func(t *testing.T) {
			for i, errField := range summary.errorFields {
				// Our logging library wraps the error provided to log.NamedError
				// to customize how it gets rendered. Internally, we extract the
				// undelying error for generating the report - to emulate this
				// behaviour we must reflect for the underlying error on the
				// internal type:
				// https://github.com/sourcegraph/log/blob/f53023898988779b0dabb75fda2c5c3b8d5ae3ae/internal/encoders/encoders.go#L59-L61
				//
				// TODO: https://github.com/sourcegraph/log/issues/65
				logError, ok := errField.Interface.(error)
				assert.Truef(t, ok, "expected errField %q to carry an error", errField.Key)
				rv := reflect.Indirect(reflect.ValueOf(logError))
				v := rv.FieldByName("Source")
				err, ok := v.Interface().(error)
				assert.Truef(t, ok, "expected errField %q to carry a log error that wraps an error as 'Source'", errField.Key)

				// The undelying error should only report the original error
				// exactly without any additional context. This is used in
				// Sentry report generation, which preserves our desired
				// grouping.
				assert.Equal(t, err.Error(), failureMessage)

				// Assert that the final Sentry report includes the metadata we
				// are attaching to assist with diagnostics on this particular
				// error.
				event, _ := errors.BuildSentryReport(err)
				t.Logf("Sentry Error message for field %q:\n\n%s\n\n",
					errField.Key, event.Message)
				assert.Contains(t, event.Message, "feature_"+strconv.Itoa(i))
				assert.Contains(t, event.Message, "action_"+strconv.Itoa(i))
				assert.Contains(t, event.Message, "id_"+strconv.Itoa(i))
			}
		})
	})

	t.Run("some failed", func(t *testing.T) {
		results := []events.PublishEventResult{{
			EventID:      "asdf",
			PublishError: errors.New("oh no"),
		}, {
			EventID: "asdfasdf",
		}}

		summary := summarizePublishEventsResults(results, summarizePublishEventsResultsOpts{})
		autogold.Expect("some events in batch failed to submit").Equal(t, summary.message)
		autogold.Expect("partial_failure").Equal(t, summary.result)
		assert.Len(t, summary.errorFields, 1)
		assert.Len(t, summary.succeededEvents, 1)
		autogold.Expect([]string{"asdfasdf"}).Equal(t, summary.succeededEvents)
		assert.Len(t, summary.failedEvents, 1)
	})

	t.Run("some failed, ignore unretriable errors and context cancelled", func(t *testing.T) {
		results := []events.PublishEventResult{{
			EventID:      "retriable",
			PublishError: errors.New("oh no"),
			Retryable:    true,
		}, {
			EventID: "ok",
		}, {
			EventID:      "unretriable",
			PublishError: errors.New("unretriable error"),
			Retryable:    false,
		}, {
			EventID:      "cancelled",
			PublishError: errors.Wrap(context.Canceled, "cancel"),
			Retryable:    true,
		}}

		summary := summarizePublishEventsResults(results, summarizePublishEventsResultsOpts{
			onlyReportRetriableAsFailed: true,
		})
		autogold.Expect("some events in batch failed to submit").Equal(t, summary.message)
		autogold.Expect("partial_failure").Equal(t, summary.result)
		// non-retriable pretends to succeed
		autogold.Expect([]string{"ok", "unretriable"}).Equal(t, summary.succeededEvents)
		assert.Len(t, summary.succeededEvents, 2)
		assert.Len(t, summary.errorFields, 3, "error fields")   // all errors are included in error logs
		assert.Len(t, summary.failedEvents, 2, "failed events") // only retryable is marked as failed

		// context error is a string, not an error
		assert.Nil(t, summary.errorFields[2].Interface)
		autogold.Expect("cancel: context canceled").Equal(t, summary.errorFields[2].String)
	})

	t.Run("all succeeded", func(t *testing.T) {
		results := []events.PublishEventResult{{
			EventID: "asdf",
		}, {
			EventID: "asdfasdf",
		}}

		summary := summarizePublishEventsResults(results, summarizePublishEventsResultsOpts{})
		autogold.Expect("all events in batch submitted successfully").Equal(t, summary.message)
		autogold.Expect("success").Equal(t, summary.result)
		assert.Len(t, summary.errorFields, 0)
		assert.Len(t, summary.succeededEvents, 2)
		autogold.Expect([]string{"asdf", "asdfasdf"}).Equal(t, summary.succeededEvents)
		assert.Len(t, summary.failedEvents, 0)
	})

	t.Run("all succeeded (large set)", func(t *testing.T) {
		results := make([]events.PublishEventResult, 100)
		wantSucceeded := make(map[string]bool)
		for i := range results {
			id := strconv.Itoa(i)
			results[i] = events.PublishEventResult{EventID: id}
			wantSucceeded[id] = false
		}

		summary := summarizePublishEventsResults(results, summarizePublishEventsResultsOpts{})
		autogold.Expect("all events in batch submitted successfully").Equal(t, summary.message)
		autogold.Expect("success").Equal(t, summary.result)
		assert.Len(t, summary.errorFields, 0)
		assert.Len(t, summary.succeededEvents, 100)
		assert.Len(t, summary.failedEvents, 0)

		for _, ev := range summary.succeededEvents {
			wantSucceeded[ev] = true
		}
		for ev, succeeded := range wantSucceeded {
			if !succeeded {
				t.Logf("event %s not marked as succeeded", ev)
			}
		}
	})
}

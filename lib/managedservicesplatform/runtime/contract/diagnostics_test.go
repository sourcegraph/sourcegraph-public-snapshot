package contract

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"

	oteltracesdk "go.opentelemetry.io/otel/sdk/trace"
)

type mockServiceMetadata struct{ t *testing.T }

func (m mockServiceMetadata) Name() string    { return m.t.Name() }
func (m mockServiceMetadata) Version() string { return "mock-version" }

func TestJobExecutionCheckIn(t *testing.T) {
	sentryEnv := fmt.Sprintf("%s_SENTRY_DSN", t.Name())
	sentryDSN := os.Getenv(sentryEnv)
	if sentryDSN == "" {
		t.Skipf("Skipping test, %q is not set", sentryEnv)
	}

	t.Log("This test creates a Sentry monitor, it must be deleted by hand")

	c := diagnosticsContract{
		sentryDSN:    &sentryDSN,
		cronSchedule: pointers.Ptr("8 * * * *"),
		cronDeadline: pointers.Ptr(24 * time.Hour),
		internal: internalContract{
			service:       mockServiceMetadata{t},
			logger:        logtest.Scoped(t),
			environmentID: fmt.Sprintf("test-%d", time.Now().Minute()),
		},
	}

	for _, failed := range []bool{true, false} {
		t.Run(fmt.Sprintf("failed=%v", failed), func(t *testing.T) {
			// Do not use noop provider, so that the trace ID is not zero.
			ctx, span := oteltracesdk.NewTracerProvider().
				Tracer(t.Name()).
				Start(context.Background(), "test")
			t.Cleanup(func() { span.End() })

			_, done, err := c.JobExecutionCheckIn(ctx)
			assert.NoError(t, err)

			time.Sleep(100 * time.Millisecond) // emulate some work

			if failed {
				done(errors.New("failed"))
			} else {
				done(nil)
			}
		})
	}
}

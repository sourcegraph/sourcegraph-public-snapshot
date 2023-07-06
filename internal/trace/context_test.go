package trace

import (
	"context"
	"testing"

	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/stretchr/testify/require"
)

func TestTraceFromContext(t *testing.T) {
	t.Run("not set in context, but raw opentelemetry span is", func(t *testing.T) {
		ctx := oteltrace.ContextWithSpan(context.Background(), recordingSpan{})
		tr := FromContext(ctx)
		require.NotNil(t, tr)
	})
}

type recordingSpan struct{ oteltrace.Span }

func (r recordingSpan) IsRecording() bool { return true }

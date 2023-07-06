package trace

import (
	"context"
	"testing"

	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/stretchr/testify/require"
)

func TestTraceFromContext(t *testing.T) {
	t.Run("set in context", func(t *testing.T) {
		ctx := contextWithTrace(context.Background(), &Trace{})
		tr := FromContext(ctx)
		require.NotNil(t, tr)
	})

	t.Run("not set in context", func(t *testing.T) {
		tr := FromContext(context.Background())
		require.Nil(t, tr)
	})

	t.Run("not set in context, but raw opentelemetry span is", func(t *testing.T) {
		ctx := oteltrace.ContextWithSpan(context.Background(), recordingSpan{})
		tr := FromContext(ctx)
		require.NotNil(t, tr)
	})
}

type recordingSpan struct{ oteltrace.Span }

func (r recordingSpan) IsRecording() bool { return true }

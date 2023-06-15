package trace

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
)

func TestTrace(t *testing.T) {
	t.Run("nil trace is noop", func(t *testing.T) {
		err := errors.New("fatal")
		tr := (*Trace)(nil)
		require.NotPanics(t, func() { tr.SetAttributes(attribute.String("key", "value")) })
		require.NotPanics(t, func() { tr.AddEvent("event") })
		require.NotPanics(t, func() { tr.LazyPrintf("printer") })
		require.NotPanics(t, func() { tr.SetError(err) })
		require.NotPanics(t, func() { tr.SetErrorIfNotContext(err) })
		require.NotPanics(t, func() { tr.Finish() })
		require.NotPanics(t, func() { tr.FinishWithErr(&err) })
	})
}
